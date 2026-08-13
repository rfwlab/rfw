//go:build js && wasm

package dom

import (
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

// Ensure UpdateDOM handles nodes without attributes (e.g. comments) without panicking.
func TestUpdateDOMSkipsNonElementNodes(_ *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("div")
	root.Set("id", "root")
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	SetInnerHTML(root, "<!--old-->")
	UpdateDOM("root", "<!--new-->")
}

func TestPatchFocusedNumberInputPreservesLiveValue(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "number-input")
	root.SetHTML(`<input id="stake" type="number" value="25"><span>old</span>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("#stake")
	input.SetValue("28")
	input.Call("focus")

	patchInnerHTML(root.Value, `<root data-component-id="number-input"><input id="stake" type="number" value="25"><span>new</span></root>`)

	patched := root.Query("#stake")
	if !patched.Equal(input.Value) {
		t.Fatal("number input was replaced")
	}
	if got := patched.Val(); got != "28" {
		t.Fatalf("number input value = %q, want 28", got)
	}
	if !js.Doc().Get("activeElement").Equal(patched.Value) {
		t.Fatal("number input lost focus")
	}
	if got := root.Query("span").Text(); got != "new" {
		t.Fatalf("patched text = %q, want new", got)
	}
}

func TestUpdateDOMRecoversAndAcceptsNextUpdate(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "recover-update")
	root.SetHTML("<span>old</span>")
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	previousHook := TemplateHook
	previousPanic := OnHandlerPanic
	defer func() {
		TemplateHook = previousHook
		OnHandlerPanic = previousPanic
	}()
	recovered := 0
	OnHandlerPanic = func(any, string) { recovered++ }
	TemplateHook = func(string, string) { panic("template hook") }

	UpdateDOM("recover-update", `<root data-component-id="recover-update"><span>first</span></root>`)
	TemplateHook = nil
	UpdateDOM("recover-update", `<root data-component-id="recover-update"><span>second</span></root>`)

	if recovered != 1 {
		t.Fatalf("recovered updates = %d, want 1", recovered)
	}
	if got := root.Query("span").Text(); got != "second" {
		t.Fatalf("next update text = %q, want second", got)
	}
}
