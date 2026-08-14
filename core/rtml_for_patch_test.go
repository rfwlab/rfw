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
	dom.UpdateDOMIn(host, c.ID, c.Render())
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
	waitForRenderFlush()

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
	waitForRenderFlush()

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
	waitForRenderFlush()

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
	waitForRenderFlush()
	if n := dom.Query("[data-list]").QueryAll("li").Length(); n != 0 {
		t.Fatalf("expected no rows after clearing, got %d", n)
	}
	st.Set("items", []any{map[string]any{"label": "back"}})
	waitForRenderFlush()
	if got := dom.Query("[data-list]").QueryAll("li").Length(); got != 1 {
		t.Fatalf("expected the list to come back, got %d rows", got)
	}
}

func TestUnifiedReconcilerPreservesExplicitlyKeyedRows(t *testing.T) {
	store := state.NewStore("reconcile_keyed", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "reconcile_keyed")
	store.Set("items", []any{
		map[string]any{"id": "a", "label": "one"},
		map[string]any{"id": "b", "label": "two"},
	})

	component := mountForComponent(t, "ReconcileKeyed", []byte(`<root><ul data-list>
@for:item in store:app.reconcile_keyed.items
<li [key @prop:item.id]>@prop:item.label</li>
@endfor
</ul></root>`))
	defer component.Unmount()
	rows := dom.Query("[data-list]").QueryAll("li")
	rowA, rowB := rows.Index(0), rows.Index(1)
	if rowA.Attr("data-key") != "a" || rowB.Attr("data-key") != "b" {
		t.Fatalf("explicit row keys were not rendered: %s", dom.Query("[data-list]").HTML())
	}

	store.Set("items", []any{
		map[string]any{"id": "b", "label": "two updated"},
		map[string]any{"id": "c", "label": "three"},
	})
	waitForRenderFlush()

	rows = dom.Query("[data-list]").QueryAll("li")
	if rows.Length() != 2 || !rows.Index(0).Equal(rowB.Value) {
		t.Fatal("keyed row identity was not retained during reorder")
	}
	if rows.Index(0).Text() != "two updated" {
		t.Fatalf("retained row was not patched: %s", rows.Index(0).Text())
	}
	if rowA.Get("isConnected").Bool() {
		t.Fatal("removed keyed row remains connected")
	}
}

func TestUnifiedReconcilerKeepsFocusedInputOutsideLoop(t *testing.T) {
	store := state.NewStore("reconcile_focus", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "reconcile_focus")
	store.Set("state", "rows")
	store.Set("query", "premier ")
	store.Set("items", []any{map[string]any{"id": "a", "label": "one"}})

	component := mountForComponent(t, "ReconcileFocus", []byte(`<root>
@if:store:app.reconcile_focus.state == "rows"
<section><input id="reconcile-search" value="@store:app.reconcile_focus.query:w"><ul>
@for:item in store:app.reconcile_focus.items
<li [key @prop:item.id]>@prop:item.label</li>
@endfor
</ul></section>
@endif
</root>`))
	defer component.Unmount()
	input := dom.ByID("reconcile-search")
	input.Call("focus")
	input.Call("setSelectionRange", 4, 8)

	store.Set("state", "rows")
	store.Set("items", []any{map[string]any{"id": "a", "label": "updated"}})
	waitForRenderFlush()

	if !dom.ByID("reconcile-search").Equal(input.Value) || !dom.Doc().Get("activeElement").Equal(input.Value) {
		t.Fatal("same-branch list refresh replaced the focused input")
	}
	if input.Val() != "premier " || input.Get("selectionStart").Int() != 4 || input.Get("selectionEnd").Int() != 8 {
		t.Fatalf("focused input state changed: value=%q selection=%d:%d", input.Val(), input.Get("selectionStart").Int(), input.Get("selectionEnd").Int())
	}
}

func TestHiddenLoopWaitsForBranchMount(t *testing.T) {
	store := state.NewStore("reconcile_hidden", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "reconcile_hidden")
	store.Set("show", "no")
	store.Set("items", []any{})

	component := mountForComponent(t, "ReconcileHidden", []byte(`<root><input id="visible-control">
@if:store:app.reconcile_hidden.show == "yes"
<ul data-hidden-list>
@for:item in store:app.reconcile_hidden.items
<li [key @prop:item.id]>@prop:item.label</li>
@endfor
</ul>
@endif
</root>`))
	defer component.Unmount()
	before := component.Stats().RenderCount
	visible := dom.ByID("visible-control")

	store.Set("items", []any{map[string]any{"id": "a", "label": "latest"}})
	waitForRenderFlush()
	if component.Stats().RenderCount != before {
		t.Fatal("hidden loop scheduled a visible component render")
	}
	if !dom.ByID("visible-control").Equal(visible.Value) {
		t.Fatal("hidden loop update touched the mounted branch")
	}

	store.Set("show", "yes")
	waitForRenderFlush()
	rows := dom.Query("[data-hidden-list]").QueryAll("li")
	if rows.Length() != 1 || rows.Index(0).Text() != "latest" {
		t.Fatalf("mounted branch did not use latest loop state: %s", dom.Query("[data-hidden-list]").HTML())
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
	waitForRenderFlush()
	oldRow := dom.Query("[data-row]")
	st.Set("items", []any{})
	waitForRenderFlush()

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

func TestStoreInputDefersWritesDuringIMEComposition(t *testing.T) {
	store := state.NewStore("ime_binding", state.WithModule("app"))
	defer state.GlobalStoreManager.UnregisterStore("app", "ime_binding")
	store.Set("value", "before")

	component := mountForComponent(t, "IMEBinding", []byte(`<root><input data-ime value="@store:app.ime_binding.value:w"></root>`))
	defer component.Unmount()
	input := dom.Query("[data-ime]")

	input.Set("value", "partial")
	composingOptions := js.NewDict()
	composingOptions.Set("bubbles", true)
	composingOptions.Set("isComposing", true)
	input.Call("dispatchEvent", js.Global().Get("InputEvent").New("input", composingOptions.Value))
	time.Sleep(20 * time.Millisecond)
	if got := store.Get("value"); got != "before" {
		t.Fatalf("composing input wrote store value %q, want before", got)
	}

	input.Set("value", "final")
	inputOptions := js.NewDict()
	inputOptions.Set("bubbles", true)
	input.Call("dispatchEvent", js.Global().Get("InputEvent").New("input", inputOptions.Value))
	time.Sleep(20 * time.Millisecond)
	if got := store.Get("value"); got != "final" {
		t.Fatalf("completed input wrote store value %q, want final", got)
	}
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
	waitForRenderFlush()

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
