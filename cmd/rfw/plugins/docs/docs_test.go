package docs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBuildAndShouldRebuild ensures the docs plugin copies files from the
// source directory to the destination and correctly reports rebuild needs.
func TestBuildAndShouldRebuild(t *testing.T) {
	p := &plugin{}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "articles")
	dest := filepath.Join(tmp, "out")
	if err := os.MkdirAll(filepath.Join(src, "a"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	srcFile := filepath.Join(src, "a", "doc.txt")
	if err := os.WriteFile(srcFile, []byte("hello"), 0o644); err != nil {
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

	// File should be copied under dest/<basename>/a/doc.txt
	copied := filepath.Join(dest, filepath.Base(src), "a", "doc.txt")
	if data, err := os.ReadFile(copied); err != nil || string(data) != "hello" {
		t.Fatalf("expected copied file, got %v %q", err, data)
	}

	if !p.ShouldRebuild(srcFile) {
		t.Fatalf("expected ShouldRebuild true for %s", srcFile)
	}
	if p.ShouldRebuild(filepath.Join(tmp, "other.txt")) {
		t.Fatalf("unexpected rebuild for unrelated file")
	}
}
