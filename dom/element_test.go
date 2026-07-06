//go:build js && wasm

package dom

import "testing"

func TestElementAttrsAndStyle(t *testing.T) {
	doc := Doc()
	el := doc.CreateElement("div")
	el.SetAttr("data-x", "y")
	if got := el.Attr("data-x"); got != "y" {
		t.Fatalf("Attr() = %q", got)
	}
	el.SetHTML("<span>ok</span>")
	if got := el.HTML(); got != "<span>ok</span>" {
		t.Fatalf("HTML() = %q", got)
	}
	el.SetStyle("color", "red")
	if v := el.Get("style").Call("getPropertyValue", "color").String(); v != "red" {
		t.Fatalf("style color = %q", v)
	}
}

func TestElementCollections(t *testing.T) {
	doc := Doc()
	parent := doc.CreateElement("div")
	parent.SetHTML("<span>a</span><span>b</span>")
	spans := parent.QueryAll("span")
	if spans.Length() != 2 {
		t.Fatalf("Length() = %d", spans.Length())
	}
	second := spans.Index(1)
	if second.Text() != "b" {
		t.Fatalf("Index(1).Text() = %q", second.Text())
	}
	second.ToggleClass("x")
	if !second.HasClass("x") {
		t.Fatalf("ToggleClass/HasClass failed")
	}
}

func TestElementAppendChild(t *testing.T) {
	doc := Doc()
	parent := doc.CreateElement("div")
	child := doc.CreateElement("span")
	parent.AppendChild(child)
	if got := parent.Query("span"); !got.Truthy() {
		t.Fatalf("AppendChild() did not append")
	}
}
