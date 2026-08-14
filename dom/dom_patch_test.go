//go:build js && wasm

package dom

import (
	"bytes"
	"log"
	"strings"
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

func TestDeferredFocusRestoreSurvivesLaterPatchPhase(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "search-input")
	root.SetHTML(`<input id="search" type="search" value="china"><span>old</span>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("#search")
	input.SetValue("china")
	input.Call("focus")
	input.Call("setSelectionRange", 5, 5)

	patchInnerHTML(root.Value, `<root data-component-id="search-input"><input id="search" type="search" value="china"><span>new</span></root>`)
	input.Call("blur")
	patchInnerHTML(root.Value, `<root data-component-id="search-input"><input id="search" type="search" value="china"><span>latest</span></root>`)

	patched := root.Query("#search")
	if !js.Doc().Get("activeElement").Equal(patched.Value) {
		t.Fatal("search input was not refocused after a later patch phase")
	}
	if got := patched.Get("selectionStart").Int(); got != 5 {
		t.Fatalf("selection start = %d, want 5", got)
	}
}

func TestRememberedFocusSurvivesRenderWorkBeforePatch(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "early-render")
	root.SetHTML(`<input id="early-search" type="search" value="china"><span>old</span>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("#early-search")
	input.SetValue("china")
	input.Call("focus")
	input.Call("setSelectionRange", 5, 5)
	input.Call("dispatchEvent", js.Global().Get("Event").New("input", map[string]any{"bubbles": true}))
	input.Call("blur")

	patchInnerHTML(root.Value, `<root data-component-id="early-render"><input id="early-search" type="search" value="china"><span>new</span></root>`)

	patched := root.Query("#early-search")
	if !js.Doc().Get("activeElement").Equal(patched.Value) {
		t.Fatal("remembered search input was not refocused")
	}
	if got := patched.Get("selectionStart").Int(); got != 5 {
		t.Fatalf("selection start = %d, want 5", got)
	}
}

func TestRememberedFocusTracksCaretMovementBeforeRefresh(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "caret-refresh")
	root.SetHTML(`<input id="caret-search" type="search" value="galway"><span>old</span>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("#caret-search")
	input.Call("focus")
	input.Call("setSelectionRange", 6, 6)
	input.Call("dispatchEvent", js.Global().Get("Event").New("input", map[string]any{"bubbles": true}))
	input.Call("setSelectionRange", 2, 2)
	js.Doc().Call("dispatchEvent", js.Global().Get("Event").New("selectionchange"))
	input.Call("blur")

	patchInnerHTML(root.Value, `<root data-component-id="caret-refresh"><input id="caret-search" type="search" value="galway"><span>new</span></root>`)

	patched := root.Query("#caret-search")
	if !js.Doc().Get("activeElement").Equal(patched.Value) {
		t.Fatal("search input was not refocused after refresh")
	}
	if got := patched.Get("selectionStart").Int(); got != 2 {
		t.Fatalf("selection start = %d, want 2", got)
	}
}

func TestDeferredFocusRestoreDoesNotStealAnotherControl(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "focus-switch")
	root.SetHTML(`<input id="search" type="search"><button id="apply">Apply</button>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	search := root.Query("#search")
	search.Call("focus")
	patchInnerHTML(root.Value, `<root data-component-id="focus-switch"><input id="search" type="search"><button id="apply">Apply</button></root>`)
	apply := root.Query("#apply")
	apply.Call("focus")
	restoreDeferredInputFocus()

	if !js.Doc().Get("activeElement").Equal(apply.Value) {
		t.Fatal("deferred restore stole focus from another control")
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

func TestDOMReporterPanicLogsOriginalFailure(t *testing.T) {
	previousPanic := OnHandlerPanic
	previousWriter := log.Writer()
	var output bytes.Buffer
	OnHandlerPanic = func(any, string) { panic("reporter failure") }
	log.SetOutput(&output)
	defer func() {
		OnHandlerPanic = previousPanic
		log.SetOutput(previousWriter)
	}()

	func() {
		defer recoverDOMUpdate("broken-component")
		panic("original failure")
	}()

	logs := output.String()
	if !strings.Contains(logs, "DOM panic reporter failed: reporter failure") {
		t.Fatalf("missing reporter failure log: %s", logs)
	}
	if !strings.Contains(logs, "recovered DOM update panic for broken-component: original failure") {
		t.Fatalf("missing original failure log: %s", logs)
	}
}
