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
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(oldwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

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
	index, err := os.ReadFile(filepath.Join(projDir, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if strings.Contains(string(index), "Date.now()") ||
		!strings.Contains(string(index), "RFW_WASM_VERSION") {
		t.Fatalf("index.html does not use the generated wasm version")
	}
	loader, err := os.ReadFile(filepath.Join(projDir, "wasm_loader.js"))
	if err != nil {
		t.Fatalf("read wasm_loader.js: %v", err)
	}
	if !strings.Contains(string(loader), "instantiateStreaming") {
		t.Fatalf("wasm_loader.js does not use streaming instantiation")
	}
}

// TestInitProjectErrors checks basic error paths.
func TestInitProjectErrors(t *testing.T) {
	if err := InitProject("", true); err == nil {
		t.Fatalf("expected error for empty project name")
	}

	dir := t.TempDir()
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)
	os.Chdir(dir)
	os.Mkdir("exists", 0755)
	if err := InitProject("exists", true); err == nil {
		t.Fatalf("expected error for existing directory")
	}
}
