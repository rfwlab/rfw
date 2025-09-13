package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreAndPostBuild verifies that the env plugin generates a temporary
// package exposing RFW_ environment variables and cleans it up afterwards.
func TestPreAndPostBuild(t *testing.T) {
	p := &plugin{}

	// Prepare environment variables.
	t.Setenv("RFW_FOO", "bar")
	t.Setenv("RFW_BAR", "baz")

	dir := t.TempDir()
	origWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	if err := p.PreBuild(nil); err != nil {
		t.Fatalf("PreBuild: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("rfwenv", "rfw_env.go"))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"BAR": "baz"`) || !strings.Contains(content, `"FOO": "bar"`) {
		t.Fatalf("generated file missing variables: %s", content)
	}

	if err := p.PostBuild(nil); err != nil {
		t.Fatalf("PostBuild: %v", err)
	}
	if _, err := os.Stat("rfwenv"); !os.IsNotExist(err) {
		t.Fatalf("rfwenv directory should be removed")
	}

	if p.ShouldRebuild("anything") {
		t.Fatalf("env plugin should never trigger rebuilds")
	}
}
