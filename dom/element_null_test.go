//go:build js && wasm

package dom

import "testing"

// A null element (missing query result) must be inert: mutators no-op and
// readers return zero values instead of panicking.
func TestNullElementIsInert(t *testing.T) {
	el := Doc().Query("#does-not-exist")
	el.SetHTML("<b>x</b>")
	el.SetText("x")
	el.SetAttr("a", "b")
	el.SetStyle("color", "red")
	el.SetValue("v")
	el.AddClass("c")
	el.RemoveClass("c")
	el.ToggleClass("c")
	if el.Text() != "" || el.HTML() != "" || el.Val() != "" || el.Attr("a") != "" ||
		el.Checked() || el.HasClass("c") || el.Data("x") != "" {
		t.Fatalf("null element readers must return zero values")
	}
}
