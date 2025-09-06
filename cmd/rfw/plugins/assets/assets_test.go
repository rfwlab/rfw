package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBuildAndShouldRebuild verifies that the assets plugin copies files and
// reports rebuild requirements for files within the source directory.
func TestBuildAndShouldRebuild(t *testing.T) {
	p := &plugin{}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "assets")
	dest := filepath.Join(tmp, "dist")
	if err := os.MkdirAll(filepath.Join(src, "img"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	srcFile := filepath.Join(src, "img", "logo.png")
	if err := os.WriteFile(srcFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{Dir: src, Dest: dest}
	raw, _ := json.Marshal(cfg)
	if err := p.Build(raw); err != nil {
		t.Fatalf("Build: %v", err)
	}

	copied := filepath.Join(dest, "img", "logo.png")
	if data, err := os.ReadFile(copied); err != nil || string(data) != "data" {
		t.Fatalf("expected copied file, got %v %q", err, data)
	}

	if !p.ShouldRebuild(srcFile) {
		t.Fatalf("expected ShouldRebuild true for %s", srcFile)
	}
	if p.ShouldRebuild(filepath.Join(tmp, "other")) {
		t.Fatalf("unexpected rebuild for unrelated file")
	}
}
