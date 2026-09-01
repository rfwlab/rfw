//go:build js && wasm

package core

import (
	"testing"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
)

func mountForComponent(t *testing.T, name string, tpl []byte) *HTMLComponent {
	t.Helper()
	host := dom.Doc().CreateElement("div")
	dom.Doc().Body().AppendChild(host)
	t.Cleanup(func() { host.Call("remove") })

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

func TestForPatchRebindsInputsOnce(t *testing.T) {
	st := state.NewStore("forpatch3", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatch3")
	st.Set("name", "")
	st.Set("ast", "")
	st.Set("row", "")
	st.Set("items", []any{map[string]any{"label": "one"}})

	tpl := []byte(`<root><input data-name value="@store:app.forpatch3.name:w"><input data-ast data-bind-store="app.forpatch3.ast"><ul>@for:it in store:app.forpatch3.items <li><input data-row data-bind-store="app.forpatch3.row">@prop:it.label</li>@endfor</ul></root>`)
	c := mountForComponent(t, "ForPatch3", tpl)
	defer c.Unmount()

	st.Set("items", []any{map[string]any{"label": "two"}})
	oldRow := dom.Query("[data-row]")
	st.Set("items", []any{})

	oldHook := state.StoreHook
	defer func() { state.StoreHook = oldHook }()
	sets := make(chan string, 4)
	state.StoreHook = func(module, store, key string, value any) {
		if module == "app" && store == "forpatch3" {
			sets <- key
		}
		if oldHook != nil {
			oldHook(module, store, key, value)
		}
	}

	oldRow.Value.Set("value", "detached")
	oldRow.Value.Call("dispatchEvent", js.CustomEvent().New("input"))
	select {
	case key := <-sets:
		t.Fatalf("detached row updated store key %q", key)
	case <-time.After(20 * time.Millisecond):
	}

	input := dom.Query("[data-name]")
	input.Value.Set("value", "Mirko")
	input.Value.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneStoreSet(t, sets, "name")

	input = dom.Query("[data-ast]")
	input.Value.Set("value", "AST")
	input.Value.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneStoreSet(t, sets, "ast")
}

func expectOneStoreSet(t *testing.T, sets <-chan string, want string) {
	t.Helper()
	select {
	case got := <-sets:
		if got != want {
			t.Fatalf("input updated store key %q, want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("input did not update the store")
	}
	select {
	case got := <-sets:
		t.Fatalf("input updated store key %q more than once", got)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestForPatchRebindsSignalInputsOnce(t *testing.T) {
	st := state.NewStore("forpatch4", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatch4")
	st.Set("items", []any{map[string]any{"label": "one"}})
	legacy := state.NewSignal("legacy")
	ast := state.NewSignal("ast")

	host := dom.Doc().CreateElement("div")
	dom.Doc().Body().AppendChild(host)
	t.Cleanup(func() { host.Call("remove") })
	tpl := []byte(`<root><input data-legacy value="@signal:legacy:w"><span hidden>@signal:ast</span><input data-ast-signal data-bind-signal="ast"><ul>@for:it in store:app.forpatch4.items <li>@prop:it.label</li>@endfor</ul></root>`)
	c := NewHTMLComponent("ForPatch4", tpl, map[string]any{
		"legacy": legacy,
		"ast":    ast,
	})
	c.SetComponent(c)
	c.Init(nil)
	host.SetHTML(c.Render())
	c.Mount()
	defer c.Unmount()

	st.Set("items", []any{map[string]any{"label": "two"}})
	st.Set("items", []any{map[string]any{"label": "three"}})

	legacySets := make(chan string, 2)
	legacySub := legacy.OnChange(func(value string) { legacySets <- value })
	defer legacySub.Stop()
	astSets := make(chan string, 2)
	astSub := ast.OnChange(func(value string) { astSets <- value })
	defer astSub.Stop()

	input := dom.Query("[data-legacy]")
	input.Value.Set("value", "legacy-updated")
	input.Value.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneSignalSet(t, legacySets, "legacy-updated")

	input = dom.Query("[data-ast-signal]")
	input.Value.Set("value", "ast-updated")
	input.Value.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneSignalSet(t, astSets, "ast-updated")
}

func expectOneSignalSet(t *testing.T, sets <-chan string, want string) {
	t.Helper()
	select {
	case got := <-sets:
		if got != want {
			t.Fatalf("signal value = %q, want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("input did not update the signal")
	}
	select {
	case <-sets:
		t.Fatal("input updated the signal more than once")
	case <-time.After(20 * time.Millisecond):
	}
}
