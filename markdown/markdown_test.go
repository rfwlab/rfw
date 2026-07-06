package markdown

import (
	"strings"
	"testing"
)

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

// TestParsePreservesQuotes ensures markdown.Parse keeps plain quotes so
// component include directives remain parseable.
func TestParsePreservesQuotes(t *testing.T) {
	src := `@include:Comp:{code:"/path/file.go", uri:"/demo"}`
	html := Parse(src)
	if !strings.Contains(html, `code:&quot;/path/file.go&quot;`) {
		t.Fatalf("expected quoted code path, got %q", html)
	}
	if strings.Contains(html, "&ldquo;") || strings.Contains(html, "&rdquo;") {
		t.Fatalf("unexpected smart quotes in %q", html)
	}
}
