package markdown

import "testing"

func TestParse(t *testing.T) {
	md := "# Title"
	html := Parse(md)
	if html == "" || html[:4] != "<h1>" {
		t.Fatalf("unexpected html: %q", html)
	}
}

func TestHeadings(t *testing.T) {
	md := "# Title\n## Sub"
	hs := Headings(md)
	if len(hs) != 2 {
		t.Fatalf("expected 2 headings, got %d", len(hs))
	}
	if hs[0].Text != "Title" || hs[0].Depth != 1 {
		t.Fatalf("unexpected first heading: %+v", hs[0])
	}
}
