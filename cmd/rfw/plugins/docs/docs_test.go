//go:build !js

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
	t.Chdir(tmp)
	src := "articles"
	dest := "out"
	if err := os.MkdirAll(filepath.Join(src, "a"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	srcFile := filepath.Join(src, "a", "doc.txt")
	if err := os.WriteFile(srcFile, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	cfg := struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{Dir: src, Dest: dest}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Build(raw); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// File should be copied under dest/<basename>/a/doc.txt
	if data, err := os.ReadFile("out/articles/a/doc.txt"); err != nil || string(data) != "hello" {
		t.Fatalf("expected copied file, got %v %q", err, data)
	}

	if !p.ShouldRebuild(srcFile) {
		t.Fatalf("expected ShouldRebuild true for %s", srcFile)
	}
	if p.ShouldRebuild(filepath.Join(tmp, "other.txt")) {
		t.Fatalf("unexpected rebuild for unrelated file")
	}
}

func TestBuildRejectsPathsOutsideProject(t *testing.T) {
	p := &plugin{}
	raw, err := json.Marshal(struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{Dir: "../articles", Dest: "build/static"})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Build(raw); err == nil {
		t.Fatal("expected source outside project to be rejected")
	}
}
