//go:build !js

package build

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v2/cmd/rfw/initproj"
)

// A fresh scaffold must resolve the CLI-injected SSC import during tidy, so
// its first build can use a locked module graph without another network fetch.
func TestScaffoldBuildsOfflineAfterTidy(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if err := initproj.InitProject("example.com/scaffoldcheck", true); err != nil {
		t.Fatal(err)
	}
	t.Chdir("scaffoldcheck")
	edit := exec.Command("go", "mod", "edit", "-require=github.com/rfwlab/rfw/v2@v2.0.0", "-replace=github.com/rfwlab/rfw/v2="+root)
	if out, err := edit.CombinedOutput(); err != nil {
		t.Fatalf("configure local framework: %v\n%s", err, out)
	}
	if out, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		t.Fatalf("tidy scaffold: %v\n%s", err, out)
	}
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	sum, err := os.ReadFile("go.sum")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOFLAGS", "-mod=readonly")
	// Compression has separate coverage; avoid racing Brotli over a full bundle.
	t.Setenv("RFW_DEV_BUILD", "1")
	if err := Build(); err != nil {
		t.Fatalf("build fresh scaffold: %v", err)
	}
	wasm, err := os.ReadFile(filepath.Join("build", "client", "app.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	if len(wasm) == 0 {
		t.Fatal("scaffold wasm is empty")
	}
	if _, err := os.Stat(filepath.Join("build", "host", "host")); err != nil {
		t.Fatal(err)
	}
	if err := Build(); err != nil {
		t.Fatalf("repeat offline build: %v", err)
	}
	again, err := os.ReadFile(filepath.Join("build", "client", "app.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wasm, again) {
		t.Fatal("unchanged scaffold produced different wasm")
	}
	afterMod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	afterSum, err := os.ReadFile("go.sum")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mod, afterMod) || !bytes.Equal(sum, afterSum) {
		t.Fatal("build changed the scaffold module lock")
	}
	deps := exec.Command("go", "list", "-deps", ".")
	deps.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	out, err := deps.CombinedOutput()
	if err != nil {
		t.Fatalf("list static source dependencies: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "github.com/rfwlab/rfw/v2/hostclient") {
		t.Fatal("dependency anchor links hostclient into a normal static build")
	}
}
