//go:build !js

package initproj

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitProjectSuccess verifies project scaffolding.
func TestInitProjectSuccess(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := InitProject("example.com/testproj", true); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	projDir := filepath.Join(dir, "testproj")
	if _, err := os.Stat(filepath.Join(projDir, "go.mod")); err != nil {
		t.Fatalf("go.mod not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projDir, "wasm_exec.js")); err != nil {
		t.Fatalf("wasm_exec.js not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projDir, "wasm_loader.js")); err != nil {
		t.Fatalf("wasm_loader.js not created: %v", err)
	}
	projectRoot, err := os.OpenRoot(projDir)
	if err != nil {
		t.Fatalf("open project root: %v", err)
	}
	t.Cleanup(func() {
		if err := projectRoot.Close(); err != nil {
			t.Errorf("close project root: %v", err)
		}
	})
	index, err := projectRoot.ReadFile("index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if strings.Contains(string(index), "Date.now()") ||
		!strings.Contains(string(index), "RFW_WASM_VERSION") {
		t.Fatalf("index.html does not use the generated wasm version")
	}
	loader, err := projectRoot.ReadFile("wasm_loader.js")
	if err != nil {
		t.Fatalf("read wasm_loader.js: %v", err)
	}
	if !strings.Contains(string(loader), "instantiateStreaming") {
		t.Fatalf("wasm_loader.js does not use streaming instantiation")
	}
	if !strings.Contains(string(loader), "handleRuntimeExit") ||
		!strings.Contains(string(loader), "rfw:runtime-reload") ||
		!strings.Contains(string(loader), "writeRecovery") ||
		!strings.Contains(string(loader), "reloadOnExit = true") {
		t.Fatalf("wasm_loader.js does not include bounded runtime recovery")
	}
}

// TestInitProjectErrors checks basic error paths.
func TestInitProjectErrors(t *testing.T) {
	if err := InitProject("", true); err == nil {
		t.Fatalf("expected error for empty project name")
	}

	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("exists", 0o750); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := InitProject("exists", true); err == nil {
		t.Fatalf("expected error for existing directory")
	}
}

func TestValidateModulePath(t *testing.T) {
	for _, modulePath := range []string{"example", "example.com/team/app", "example.com/team/app/v2"} {
		if err := validateModulePath(modulePath); err != nil {
			t.Errorf("expected %q to be valid: %v", modulePath, err)
		}
	}
	for _, modulePath := range []string{
		"",
		"-x",
		"../app",
		"example.com//app",
		"example.com/app name",
		"example.com/app\nreplace x",
		"example.com/app+tools",
		"example.com/.hidden",
		"example.com/NUL.txt",
	} {
		if err := validateModulePath(modulePath); err == nil {
			t.Errorf("expected %q to be rejected", modulePath)
		}
	}
}

func TestInitProjectPinsSelectedVersionWithoutNetwork(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GOPROXY", "off")
	const selected = "v2.3.1-0.20261016153501-8ce9ba2d776b"
	if err := InitProjectWithVersion("example.com/pinned", true, selected); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("pinned/go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "require github.com/rfwlab/rfw/v2 "+selected) || strings.Contains(string(data), "replace") {
		t.Fatalf("generated module: %s", data)
	}
}

func TestInitProjectRequiresVersionBeforeCreatingDevelopmentProject(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := InitProjectWithVersion("example.com/unversioned", false, "(devel)"); err == nil {
		t.Fatal("unversioned CLI initialized an unpinned project")
	}
	if _, err := os.Stat("unversioned"); !os.IsNotExist(err) {
		t.Fatalf("unversioned project directory: %v", err)
	}
}

func TestInitProjectRejectsInvalidVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GOPROXY", "off")
	if err := InitProjectWithVersion("example.com/invalidversion", true, "v2.3.0\nreplace example.com/other => ../other"); err == nil {
		t.Fatal("invalid framework version accepted")
	}
}
