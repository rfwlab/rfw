//go:build !js

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
	t.Chdir(tmp)
	src := "assets"
	dest := "dist"
	if err := os.MkdirAll(filepath.Join(src, "img"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	srcFile := filepath.Join(src, "img", "logo.png")
	if err := os.WriteFile(srcFile, []byte("data"), 0o600); err != nil {
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

	if data, err := os.ReadFile("dist/img/logo.png"); err != nil || string(data) != "data" {
		t.Fatalf("expected copied file, got %v %q", err, data)
	}

	if !p.ShouldRebuild(srcFile) {
		t.Fatalf("expected ShouldRebuild true for %s", srcFile)
	}
	if p.ShouldRebuild(filepath.Join(tmp, "other")) {
		t.Fatalf("unexpected rebuild for unrelated file")
	}
}

func TestBuildRejectsPathsOutsideProject(t *testing.T) {
	p := &plugin{}
	for _, cfg := range []struct {
		Dir  string `json:"dir"`
		Dest string `json:"dest"`
	}{
		{Dir: "../assets", Dest: "dist"},
		{Dir: "assets", Dest: "../dist"},
	} {
		raw, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Build(raw); err == nil {
			t.Errorf("expected config %+v to be rejected", cfg)
		}
	}
}
