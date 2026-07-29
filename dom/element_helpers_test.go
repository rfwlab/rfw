//go:build js && wasm

package dom

import "testing"

func TestElementRemoveAttr(t *testing.T) {
	el := Doc().CreateElement("button")
	el.SetAttr("disabled", "")
	if !el.Call("hasAttribute", "disabled").Bool() {
		t.Fatal("SetAttr did not set disabled")
	}
	el.RemoveAttr("disabled")
	if el.Call("hasAttribute", "disabled").Bool() {
		t.Fatal("RemoveAttr left the attribute in place")
	}
}

func TestElementMatches(t *testing.T) {
	el := Doc().CreateElement("div")
	el.SetAttr("data-row", "1")
	if !el.Matches("[data-row]") {
		t.Fatal("Matches() = false for a matching selector")
	}
	if el.Matches("[data-other]") {
		t.Fatal("Matches() = true for a non-matching selector")
	}
}

func TestMissingElementHelpersAreSafe(t *testing.T) {
	el := Query("#definitely-not-in-the-document")
	el.RemoveAttr("disabled")
	if el.Matches("[data-row]") {
		t.Fatal("Matches() = true on a missing element")
	}
}

func TestFromWrapsRawValue(t *testing.T) {
	parent := Doc().CreateElement("div")
	parent.SetHTML(`<span data-id="7" class="a">x</span>`)
	raw := parent.Call("querySelector", "span")

	el := From(raw)
	if got := el.Data("id"); got != "7" {
		t.Fatalf("Data(id) = %q", got)
	}
	if !el.HasClass("a") {
		t.Fatal("HasClass(a) = false")
	}
	if c := el.Closest("div"); c.IsNull() {
		t.Fatal("Closest(div) returned null")
	}
}
