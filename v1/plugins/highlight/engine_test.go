package highlight

import (
	"strings"
	"testing"
)

func TestHighlightGo(t *testing.T) {
	out, ok := Highlight("package main", "go")
	if !ok || !strings.Contains(out, `<span class="hl-kw">package</span>`) {
		t.Fatalf("missing keyword highlighting: %s", out)
	}
}

func TestHighlightRTML(t *testing.T) {
	code := "<div class=\"x\">{name}</div>\n<button @on:click:save>ok</button>\n@for:item in items"
	out, ok := Highlight(code, "rtml")
	if !ok {
		t.Fatalf("unexpected highlight failure: %s", out)
	}
	tag := "<span class=\"hl-tag\">&lt;div <span class=\"hl-attr\">class</span>=<span class=\"hl-string\">&#34;x&#34;</span>&gt;</span>"
	if !strings.Contains(out, tag) {
		t.Fatalf("missing tag highlighting: %s", out)
	}
	if !strings.Contains(out, `<span class="hl-var">{name}</span>`) {
		t.Fatalf("missing variable highlighting: %s", out)
	}
	if !strings.Contains(out, `<span class="hl-cmd">@on:click:save</span>`) {
		t.Fatalf("missing command attribute highlighting: %s", out)
	}
	if !strings.Contains(out, `<span class="hl-cmd">@for:item</span>`) {
		t.Fatalf("missing standalone command highlighting: %s", out)
	}
}
