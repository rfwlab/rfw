//go:build !js

package copy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBuildAndShouldRebuild ensures files matching patterns are copied
// and rebuilds are triggered for matched paths.
func TestBuildAndShouldRebuild(t *testing.T) {
	p := &plugin{}
	tmp := t.TempDir()

	t.Chdir(tmp)

	srcRoot := filepath.Join("examples", "components")
	if err := os.MkdirAll(filepath.Join(srcRoot, "templates"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	fileA := filepath.Join(srcRoot, "comp.txt")
	if err := os.WriteFile(fileA, []byte("a"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	fileB := filepath.Join(srcRoot, "templates", "tpl.txt")
	if err := os.WriteFile(fileB, []byte("b"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	destRoot := filepath.Join("build", "static", "examples", "components")
	cfg := struct {
		Files []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"files"`
	}{
		Files: []struct {
			From string `json:"from"`
			To   string `json:"to"`
		}{{
			From: filepath.Join(srcRoot, "**", "*"),
			To:   destRoot,
		}},
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Build(raw); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if data, err := os.ReadFile("build/static/examples/components/comp.txt"); err != nil || string(data) != "a" {
		t.Fatalf("comp.txt not copied: %v %s", err, data)
	}
	if data, err := os.ReadFile("build/static/examples/components/templates/tpl.txt"); err != nil || string(data) != "b" {
		t.Fatalf("tpl.txt not copied: %v %s", err, data)
	}

	if !p.ShouldRebuild(fileA) {
		t.Fatalf("expected ShouldRebuild true for %s", fileA)
	}
	if p.ShouldRebuild(filepath.Join("other.txt")) {
		t.Fatalf("unexpected rebuild for unrelated file")
	}
}

func TestBuildRejectsPathsOutsideProject(t *testing.T) {
	p := &plugin{}
	for _, copyRule := range []rule{
		{From: "../*", To: "build"},
		{From: "*", To: "../build"},
		{From: "*", To: filepath.Join(t.TempDir(), "build")},
	} {
		raw, err := json.Marshal(struct {
			Files []rule `json:"files"`
		}{Files: []rule{copyRule}})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Build(raw); err == nil {
			t.Errorf("expected rule %+v to be rejected", copyRule)
		}
	}
}
