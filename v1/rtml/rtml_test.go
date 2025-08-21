package rtml

import (
	"strings"
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

type dummyDep struct{ content string }

func (d dummyDep) Render() string { return d.content }

func TestReplaceForLoops(t *testing.T) {
	items := []any{dummyDep{"<p>one</p>"}, dummyDep{"<p>two</p>"}}
	tpl := "<root>@for:item in items @prop:item @endfor</root>"
	ctx := Context{Props: map[string]any{"items": items}}
	out := Replace(tpl, ctx)
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") {
		t.Fatalf("expected loop to render items, got %s", out)
	}
}

func TestReplaceConditionals(t *testing.T) {
	tpl := `@if:prop:val=="yes"
Ok
@else
No
@endif`
	ctx := Context{Props: map[string]any{"val": "no"}}
	out := Replace(tpl, ctx)
	if strings.Contains(out, "Ok") || !strings.Contains(out, "No") {
		t.Fatalf("conditional not evaluated correctly: %s", out)
	}
}

func TestReplaceStorePlaceholders(t *testing.T) {
	store := state.NewStore("default", state.WithModule("app"))
	store.Set("count", 5)
	tpl := `<p>Count: @store:app.default.count</p><input value="@store:app.default.count:w">`
	ctx := Context{}
	out := Replace(tpl, ctx)
	if !strings.Contains(out, "Count: 5") {
		t.Fatalf("expected store value replacement: %s", out)
	}
	if !strings.Contains(out, "@store:app.default.count:w") {
		t.Fatalf("writable store placeholder replaced unexpectedly: %s", out)
	}
}

func TestReplaceRtIsAttributes(t *testing.T) {
	tpl := `<div rt-is="child"></div>`
	ctx := Context{Dependencies: map[string]Dependency{"child": dummyDep{"<span>child</span>"}}}
	out := Replace(tpl, ctx)
	if !strings.Contains(out, "<span>child</span>") {
		t.Fatalf("expected rt-is component rendered: %s", out)
	}
}
