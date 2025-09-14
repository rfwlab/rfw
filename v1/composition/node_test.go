//go:build js && wasm

package composition

import (
	"testing"

	"github.com/rfwlab/rfw/v1/dom"
)

func TestBind(t *testing.T) {
	doc := dom.Doc()
	doc.Body().SetHTML("<div id='root'><span>old</span></div>")

	Bind("#root", func(el El) {
		el.Clear()
		el.Append(Div().Text("new"))
	})

	root := doc.ByID("root")
	if html := root.HTML(); html != "<div>new</div>" {
		t.Fatalf("expected <div>new</div>, got %q", html)
	}
}

func TestFor(t *testing.T) {
	doc := dom.Doc()
	doc.Body().SetHTML("<div id='list'></div>")
	items := []string{"a", "b"}
	i := 0

	For("#list", func() Node {
		if i >= len(items) {
			return nil
		}
		n := Div().Text(items[i])
		i++
		return n
	})

	got := doc.ByID("list").HTML()
	if got != "<div>a</div><div>b</div>" {
		t.Fatalf("unexpected html %q", got)
	}
}

func TestDivBuilder(t *testing.T) {
	d := Div().Class("c").Style("color", "red").Text("hi")
	el := d.Element()
	if !el.HasClass("c") {
		t.Fatalf("expected class c")
	}
	if el.Text() != "hi" {
		t.Fatalf("expected text hi")
	}
	if v := el.Get("style").Call("getPropertyValue", "color").String(); v != "red" {
		t.Fatalf("expected style color red, got %q", v)
	}
}
