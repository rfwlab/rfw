package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadPortFromManifest verifies that readPortFromManifest reads the port
// value from rfw.json when present and falls back to an empty string when the
// file is missing or malformed.
func TestReadPortFromManifest(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Missing file should return empty string.
	if got := readPortFromManifest(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}

	// Create manifest with a port value.
	data := []byte(`{"port": 9090}`)
	if err := os.WriteFile(filepath.Join(dir, "rfw.json"), data, 0o644); err != nil {
		t.Fatalf("write rfw.json: %v", err)
	}
	if got := readPortFromManifest(); got != "9090" {
		t.Fatalf("expected 9090, got %q", got)
	}

	// Zero port should return empty string.
	if err := os.WriteFile("rfw.json", []byte(`{"port":0}`), 0o644); err != nil {
		t.Fatalf("rewrite rfw.json: %v", err)
	}
	if got := readPortFromManifest(); got != "" {
		t.Fatalf("expected empty string for zero port, got %q", got)
	}
}
