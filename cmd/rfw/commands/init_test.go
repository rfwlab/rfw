//go:build !js

package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/mirkobrombin/go-cli-builder/v1/root"
)

func TestInitCommandDoesNotUseProjectNameAsFrameworkVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"rfw", "init", "--skip-tidy", "example.com/defaultcli"}
	rc := root.NewRootCommand("rfw", "", "", "")
	cmd := NewInitCommand()
	rc.AddCommand(cmd)
	if err := rc.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := cmd.GetFlagString("rfw-version"); got != "" {
		t.Fatalf("implicit framework flag: %q", got)
	}
}

func TestInitCommandPinsExplicitFrameworkVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GOPROXY", "off")
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"rfw", "init", "--rfw-version", "v2.3.0", "--skip-tidy", "example.com/explicitcli"}
	rc := root.NewRootCommand("rfw", "", "", "")
	rc.AddCommand(NewInitCommand())
	if err := rc.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("explicitcli/go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "require github.com/rfwlab/rfw/v2 v2.3.0") {
		t.Fatalf("explicit CLI dependency: %s", data)
	}
}
