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
	side.Set("__marker", "kept")

	st.Set("items", []any{
		map[string]any{"label": "one"},
		map[string]any{"label": "two"},
		map[string]any{"label": "three"},
	})

	sideAfter := dom.ByID("forpatch-side")
	if sideAfter.IsNull() {
		t.Fatal("sibling vanished after the list changed")
	}
	if got := sideAfter.Get("__marker"); !got.Truthy() || got.String() != "kept" {
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

func TestForPatchPreservesKeyedRows(t *testing.T) {
	st := state.NewStore("forpatchkeyed", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatchkeyed")
	st.Set("items", []any{
		map[string]any{"id": "a", "label": "one"},
		map[string]any{"id": "b", "label": "two"},
	})

	tpl := []byte(`<root><ul data-list>@for:it in store:app.forpatchkeyed.items <li [key @prop:it.id]>@prop:it.label</li>@endfor</ul></root>`)
	mountForComponent(t, "ForPatchKeyed", tpl)

	rows := dom.Query("[data-list]").QueryAll("li")
	rowA := rows.Index(0)
	rowB := rows.Index(1)
	rowA.Set("__marker", "a")
	rowB.Set("__marker", "b")

	st.Set("items", []any{
		map[string]any{"id": "b", "label": "two updated"},
		map[string]any{"id": "c", "label": "three"},
	})

	rows = dom.Query("[data-list]").QueryAll("li")
	if rows.Length() != 2 {
		t.Fatalf("expected 2 rows, got %d", rows.Length())
	}
	if !rows.Index(0).Equal(rowB.Value) {
		t.Fatal("keyed row b was replaced during reorder")
	}
	if got := rows.Index(0).Text(); got != "two updated" {
		t.Fatalf("patched row text = %q", got)
	}
	if rowA.Get("isConnected").Bool() {
		t.Fatal("removed keyed row remains connected")
	}
	if got := rows.Index(1).Get("__marker"); got.Truthy() {
		t.Fatal("inserted row inherited another row's identity")
	}
}

func TestForPatchKeepsFocusedInputOutsideLoop(t *testing.T) {
	st := state.NewStore("forpatchfocus", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatchfocus")
	st.Set("state", "rows")
	st.Set("query", "premier ")
	st.Set("items", []any{map[string]any{"id": "a", "label": "one"}})

	tpl := []byte(`<root>
@if:store:app.forpatchfocus.state == "rows"
<section><input id="forpatch-search" value="@store:app.forpatchfocus.query:w"><ul>@for:it in store:app.forpatchfocus.items <li [key @prop:it.id]>@prop:it.label</li>@endfor</ul></section>
@endif
</root>`)
	mountForComponent(t, "ForPatchFocus", tpl)

	input := dom.ByID("forpatch-search")
	input.Call("focus")
	input.Call("setSelectionRange", 8, 8)
	st.Set("state", "rows")
	st.Set("items", []any{map[string]any{"id": "a", "label": "updated"}})

	if !dom.Doc().Get("activeElement").Equal(input.Value) {
		t.Fatal("list refresh replaced the focused input")
	}
	if got := input.Val(); got != "premier " {
		t.Fatalf("focused input value = %q", got)
	}
	if got := input.Get("selectionStart").Int(); got != 8 {
		t.Fatalf("selectionStart = %d, want 8", got)
	}
}

func TestHiddenForPatchDoesNotRenderMountedBranch(t *testing.T) {
	st := state.NewStore("forpatchhidden", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatchhidden")
	st.Set("show", "no")
	st.Set("items", []any{map[string]any{"id": "a", "label": "one"}})

	tpl := []byte(`<root><input id="forpatch-visible">
@if:store:app.forpatchhidden.show == "yes"
<ul data-hidden-list>@for:it in store:app.forpatchhidden.items <li [key @prop:it.id]>@prop:it.label</li>@endfor</ul>
@endif
</root>`)
	mountForComponent(t, "ForPatchHidden", tpl)

	oldHook := dom.TemplateHook
	defer func() { dom.TemplateHook = oldHook }()
	updates := 0
	dom.TemplateHook = func(string, string) { updates++ }

	st.Set("items", []any{map[string]any{"id": "a", "label": "latest"}})
	if updates != 0 {
		t.Fatalf("hidden loop triggered %d component renders", updates)
	}

	st.Set("show", "yes")
	rows := dom.Query("[data-hidden-list]").QueryAll("li")
	if rows.Length() != 1 || rows.Index(0).Text() != "latest" {
		t.Fatalf("revealed loop did not render latest store value: %s", dom.Query("[data-hidden-list]").HTML())
	}
}

func TestHiddenEmptyForRendersRowsWhenBranchMounts(t *testing.T) {
	st := state.NewStore("forpatchhiddenempty", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatchhiddenempty")
	st.Set("show", "no")
	st.Set("items", []any{})

	tpl := []byte(`<root>
@if:store:app.forpatchhiddenempty.show == "yes"
<ul data-hidden-list>@for:it in store:app.forpatchhiddenempty.items <li [key @prop:it.id]>@prop:it.label</li>@endfor</ul>
@endif
</root>`)
	mountForComponent(t, "ForPatchHiddenEmpty", tpl)

	st.Set("items", []any{map[string]any{"id": "a", "label": "arrived"}})
	st.Set("show", "yes")

	rows := dom.Query("[data-hidden-list]").QueryAll("li")
	if rows.Length() != 1 || rows.Index(0).Text() != "arrived" {
		t.Fatalf("revealed loop did not render rows received while hidden: %s", dom.Query("[data-hidden-list]").HTML())
	}
}

func TestMissingVisibleForAnchorFallsBackToRender(t *testing.T) {
	st := state.NewStore("forpatchmissinganchor", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "forpatchmissinganchor")
	st.Set("show", "yes")
	st.Set("items", []any{map[string]any{"id": "a", "label": "one"}})

	tpl := []byte(`<root>
@if:store:app.forpatchmissinganchor.show == "yes"
<ul data-list>@for:it in store:app.forpatchmissinganchor.items <li [key @prop:it.id]>@prop:it.label</li>@endfor</ul>
@endif
</root>`)
	mountForComponent(t, "ForPatchMissingAnchor", tpl)
	dom.Query("[data-for-anchor]").Call("remove")

	st.Set("items", []any{map[string]any{"id": "a", "label": "recovered"}})
	anchor := dom.Query("[data-for-anchor]")
	if anchor.IsNull() || anchor.IsUndefined() {
		t.Fatal("full render did not restore a missing visible anchor")
	}
	if got := dom.Query("[data-list]").Query("li").Text(); got != "recovered" {
		t.Fatalf("recovered row text = %q", got)
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

	oldRow.Set("value", "detached")
	oldRow.Call("dispatchEvent", js.CustomEvent().New("input"))
	select {
	case key := <-sets:
		t.Fatalf("detached row updated store key %q", key)
	case <-time.After(20 * time.Millisecond):
	}

	input := dom.Query("[data-name]")
	input.Set("value", "Mirko")
	input.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneStoreSet(t, sets, "name")

	input = dom.Query("[data-ast]")
	input.Set("value", "AST")
	input.Call("dispatchEvent", js.CustomEvent().New("input"))
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
	input.Set("value", "legacy-updated")
	input.Call("dispatchEvent", js.CustomEvent().New("input"))
	expectOneSignalSet(t, legacySets, "legacy-updated")

	input = dom.Query("[data-ast-signal]")
	input.Set("value", "ast-updated")
	input.Call("dispatchEvent", js.CustomEvent().New("input"))
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
