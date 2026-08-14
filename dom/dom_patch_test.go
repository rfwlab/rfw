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

func TestPatchPreservesFocusedTextInputIdentityAndCaret(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "search-input")
	root.SetHTML(`<input type="search" value="china"><span>old</span>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("input")
	input.SetValue("chinaa")
	input.Call("focus")
	input.Call("setSelectionRange", 2, 4, "forward")

	patchInnerHTML(root.Value, `<root data-component-id="search-input"><input type="search" value="server"><span>new</span></root>`)

	patched := root.Query("input")
	if !patched.Equal(input.Value) {
		t.Fatal("search input was replaced")
	}
	if !js.Doc().Get("activeElement").Equal(patched.Value) {
		t.Fatal("search input lost focus")
	}
	if got := patched.Val(); got != "chinaa" {
		t.Fatalf("live input value = %q, want chinaa", got)
	}
	if got := patched.Get("selectionStart").Int(); got != 2 {
		t.Fatalf("selection start = %d, want 2", got)
	}
	if got := patched.Get("selectionEnd").Int(); got != 4 {
		t.Fatalf("selection end = %d, want 4", got)
	}
}

func TestPatchDoesNotRestoreDeliberatelyBlurredInput(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "blurred-input")
	root.SetHTML(`<input type="search"><button>Apply</button>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	input := root.Query("input")
	input.Call("focus")
	input.Call("blur")

	patchInnerHTML(root.Value, `<root data-component-id="blurred-input"><input type="search"><button>Updated</button></root>`)

	if js.Doc().Get("activeElement").Equal(input.Value) {
		t.Fatal("patch restored focus after an explicit blur")
	}
}

func TestPatchPreservesUncontrolledFormProperties(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "form-state")
	root.SetHTML(`<textarea>initial</textarea><input type="checkbox"><input type="radio" name="choice"><select><option>A</option><option>B</option></select>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	textarea := root.Query("textarea")
	checkbox := root.Query(`input[type="checkbox"]`)
	radio := root.Query(`input[type="radio"]`)
	selectEl := root.Query("select")
	textarea.SetValue("operator text")
	checkbox.Set("checked", true)
	radio.Set("checked", true)
	selectEl.Set("selectedIndex", 1)

	patchInnerHTML(root.Value, `<root data-component-id="form-state"><textarea>server text</textarea><input type="checkbox"><input type="radio" name="choice"><select><option selected>A</option><option>B</option></select></root>`)

	if !root.Query("textarea").Equal(textarea.Value) || root.Query("textarea").Val() != "operator text" {
		t.Fatal("textarea identity or live value was not preserved")
	}
	if !root.Query(`input[type="checkbox"]`).Equal(checkbox.Value) || !checkbox.Checked() {
		t.Fatal("checkbox identity or checked state was not preserved")
	}
	if !root.Query(`input[type="radio"]`).Equal(radio.Value) || !radio.Checked() {
		t.Fatal("radio identity or checked state was not preserved")
	}
	if !root.Query("select").Equal(selectEl.Value) || selectEl.Get("selectedIndex").Int() != 1 {
		t.Fatal("select identity or selected option was not preserved")
	}
}

func TestPatchPreservesAndReordersKeyedNodes(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "keyed-list")
	root.SetHTML(`<ul><li data-key="a">A</li><li data-key="b">B</li></ul>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	a := root.Query(`[data-key="a"]`)
	b := root.Query(`[data-key="b"]`)
	patchInnerHTML(root.Value, `<root data-component-id="keyed-list"><ul><li data-key="b">B2</li><li data-key="a">A2</li><li data-key="c">C</li></ul></root>`)

	rows := root.QueryAll("li")
	if rows.Length() != 3 || !rows.Index(0).Equal(b.Value) || !rows.Index(1).Equal(a.Value) {
		t.Fatalf("keyed rows lost identity or order: %s", root.HTML())
	}
	if rows.Index(0).Text() != "B2" || rows.Index(1).Text() != "A2" {
		t.Fatalf("keyed row contents were not patched: %s", root.HTML())
	}
}

func TestInvalidPatchPlanLeavesDOMUntouched(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "atomic-plan")
	root.SetHTML(`<ul><li data-key="a">A</li><li data-key="b">B</li></ul>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	before := root.HTML()
	a := root.Query(`[data-key="a"]`)
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		patchInnerHTML(root.Value, `<root data-component-id="atomic-plan"><ul><li data-key="a">changed</li><li data-key="a">duplicate</li></ul></root>`)
	}()

	if recovered == nil {
		t.Fatal("duplicate identity did not reject the patch")
	}
	if got := root.HTML(); got != before {
		t.Fatalf("invalid plan mutated DOM: got %s, want %s", got, before)
	}
	if !root.Query(`[data-key="a"]`).Equal(a.Value) {
		t.Fatal("invalid plan replaced a node before failing")
	}
}

func TestKeyIdentityIsScopedToItsLoop(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "loop-scopes")
	root.SetHTML(`<section><i data-for="first" data-key="0">A</i><i data-for="second" data-key="0">B</i></section>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	patchInnerHTML(root.Value, `<root data-component-id="loop-scopes"><section><i data-for="first" data-key="0">A2</i><i data-for="second" data-key="0">B2</i></section></root>`)
	rows := root.QueryAll("i")
	if rows.Length() != 2 || rows.Index(0).Text() != "A2" || rows.Index(1).Text() != "B2" {
		t.Fatalf("loop-scoped keys were treated as duplicates: %s", root.HTML())
	}
}

func TestPatchRespectsNestedDOMOwnership(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "shell")
	root.SetHTML(`<span data-shell>old</span><root data-component-id="child"><div data-child>rendered</div></root><div data-router-outlet><root data-component-id="page"><div data-page>mounted</div></root></div>`)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	child := root.Query(`[data-component-id="child"]`)
	page := root.Query(`[data-component-id="page"]`)
	child.Query("[data-child]").SetHTML("imperative child")
	page.Query("[data-page]").SetHTML("imperative page")

	patchInnerHTML(root.Value, `<root data-component-id="shell"><span data-shell>new</span><root data-component-id="child"><div data-child>stale render</div></root><div data-router-outlet><p>empty render</p></div></root>`)

	if root.Query("[data-shell]").Text() != "new" {
		t.Fatal("shell-owned node was not patched")
	}
	if !root.Query(`[data-component-id="child"]`).Equal(child.Value) || child.Query("[data-child]").Text() != "imperative child" {
		t.Fatal("parent patch crossed child component ownership")
	}
	if !root.Query(`[data-component-id="page"]`).Equal(page.Value) || page.Query("[data-page]").Text() != "imperative page" {
		t.Fatal("parent patch crossed router outlet ownership")
	}
}

func TestPatchPreservesLiveAttributesUntilTemplateChangesThem(t *testing.T) {
	body := js.Doc().Get("body")
	root := CreateElement("root")
	root.SetAttr("data-component-id", "live-attributes")
	root.SetHTML(`<section class="panel" aria-expanded="false"><span>old</span></section>`)
	recordRenderedTree(root.Value)
	body.Call("appendChild", root.Value)
	defer root.Call("remove")

	panel := root.Query("section")
	panel.AddClass("open")
	panel.SetAttr("aria-expanded", "true")
	patchInnerHTML(root.Value, `<root data-component-id="live-attributes"><section class="panel" aria-expanded="false"><span>new</span></section></root>`)
	if !panel.HasClass("open") || panel.Attr("aria-expanded") != "true" {
		t.Fatalf("unchanged template attributes erased live state: %s", root.HTML())
	}
	if panel.Query("span").Text() != "new" {
		t.Fatal("attribute preservation blocked descendant patching")
	}

	patchInnerHTML(root.Value, `<root data-component-id="live-attributes"><section class="panel disabled" aria-expanded="mixed"><span>latest</span></section></root>`)
	if panel.Attr("class") != "panel disabled" || panel.Attr("aria-expanded") != "mixed" {
		t.Fatalf("changed template attributes did not take ownership: %s", root.HTML())
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
