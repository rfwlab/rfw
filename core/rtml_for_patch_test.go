//go:build js && wasm

package core

import (
	"testing"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

func mountForComponent(t *testing.T, name string, tpl []byte) *HTMLComponent {
	t.Helper()
	host := dom.Doc().CreateElement("div")
	dom.Doc().Body().AppendChild(host)

	c := NewHTMLComponent(name, tpl, nil)
	c.SetComponent(c)
	c.Init(nil)
	host.SetHTML(c.Render())
	c.Mount()
	return c
}

// A list that changes should cost its own rows, not a re-render of everything
// around it: the sibling markup keeps its node identity.
func TestForPatchLeavesSiblingsAlone(t *testing.T) {
	st := state.NewStore("forpatch", state.WithModule("app"))
	st.Set("items", []any{
		map[string]any{"label": "one"},
		map[string]any{"label": "two"},
	})

	tpl := []byte(`<root><div id="forpatch-side">side</div><ul>@for:it in store:app.forpatch.items <li>@prop:it.label</li>@endfor</ul></root>`)
	mountForComponent(t, "ForPatch", tpl)

	side := dom.ByID("forpatch-side")
	if side.IsNull() {
		t.Fatal("sibling not rendered")
	}
	side.Value.Set("__marker", "kept")

	st.Set("items", []any{
		map[string]any{"label": "one"},
		map[string]any{"label": "two"},
		map[string]any{"label": "three"},
	})

	sideAfter := dom.ByID("forpatch-side")
	if sideAfter.IsNull() {
		t.Fatal("sibling vanished after the list changed")
	}
	if got := sideAfter.Value.Get("__marker"); !got.Truthy() || got.String() != "kept" {
		t.Fatal("sibling was re-created: the whole component re-rendered")
	}
}

// The patched rows have to match what a full render would have produced.
func TestForPatchRendersEveryRow(t *testing.T) {
	st := state.NewStore("forpatch2", state.WithModule("app"))
	st.Set("items", []any{map[string]any{"label": "a"}})

	tpl := []byte(`<root><ul data-list>@for:it in store:app.forpatch2.items <li>@prop:it.label</li>@endfor</ul></root>`)
	mountForComponent(t, "ForPatch2", tpl)

	st.Set("items", []any{
		map[string]any{"label": "x"},
		map[string]any{"label": "y"},
	})

	list := dom.Query("[data-list]")
	rows := list.QueryAll("li")
	if rows.Length() != 2 {
		t.Fatalf("expected 2 rows, got %d (%s)", rows.Length(), list.HTML())
	}
	if got := rows.Index(0).Text(); got != "x" {
		t.Fatalf("first row = %q", got)
	}
	if got := rows.Index(1).Text(); got != "y" {
		t.Fatalf("second row = %q", got)
	}

	// emptying the list clears the rows and keeps the anchor for the next value
	st.Set("items", []any{})
	if n := dom.Query("[data-list]").QueryAll("li").Length(); n != 0 {
		t.Fatalf("expected no rows after clearing, got %d", n)
	}
	st.Set("items", []any{map[string]any{"label": "back"}})
	if got := dom.Query("[data-list]").QueryAll("li").Length(); got != 1 {
		t.Fatalf("expected the list to come back, got %d rows", got)
	}
}

// A body the patch cannot own on its own (an include, a nested conditional)
// falls back to the full render instead of painting something incomplete.
func TestForPatchFallsBackForRichBodies(t *testing.T) {
	if incrementalForBody(`<li>@include:child</li>`) {
		t.Fatal("include body should not be patched incrementally")
	}
	if incrementalForBody("<li>@if:prop:x\\nyes\\n@endif</li>") {
		t.Fatal("conditional body should not be patched incrementally")
	}
	if !incrementalForBody(`<li class="@prop:it.cls">@prop:it.label</li>`) {
		t.Fatal("a plain body should be patchable")
	}
}
