//go:build !js

package pages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeriveRoute(t *testing.T) {
	tests := map[string]struct{ path, comp string }{
		"index.go":           {"/", "Index"},
		"about.go":           {"/about", "About"},
		"blog/index.go":      {"/blog", "BlogIndex"},
		"posts/[id].go":      {"/posts/:id", "PostsId"},
		"[user]/settings.go": {"/:user/settings", "UserSettings"},
	}
	for in, exp := range tests {
		p, c := deriveRoute(in)
		if p != exp.path || c != exp.comp {
			t.Errorf("%s => (%s,%s), want (%s,%s)", in, p, c, exp.path, exp.comp)
		}
	}
}

// TestPreBuildMissingDirNoop verifies that a project without a pages directory
// builds cleanly: PreBuild returns nil and generates nothing. This is what makes
// default activation safe for manually-routed projects.
func TestPreBuildMissingDirNoop(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := (&plugin{}).PreBuild(nil); err != nil {
		t.Fatalf("PreBuild with no pages dir should be a no-op, got %v", err)
	}
	if _, err := os.Stat(filepath.Join("pages", "routes_gen.go")); !os.IsNotExist(err) {
		t.Fatalf("no routes file should be generated without a pages dir")
	}
}
