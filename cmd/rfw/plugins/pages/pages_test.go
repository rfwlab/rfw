package pages

import "testing"

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
