package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIncrementPort verifies port arithmetic.
func TestIncrementPort(t *testing.T) {
	if got := incrementPort("8080"); got != "8081" {
		t.Fatalf("expected 8081, got %s", got)
	}
}

// TestReadBuildType reads build type from temporary manifest.
func TestReadBuildType(t *testing.T) {
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer os.Chdir(oldwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	// No file should return empty string.
	if got := readBuildType(); got != "" {
		t.Fatalf("expected empty build type, got %q", got)
	}

	// Create manifest and verify value.
	data := []byte(`{"build":{"type":"SSC"}}`)
	if err := os.WriteFile(filepath.Join(dir, "rfw.json"), data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if got := readBuildType(); got != "ssc" {
		t.Fatalf("expected 'ssc', got %q", got)
	}
}
