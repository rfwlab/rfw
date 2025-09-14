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

func TestBindEl(t *testing.T) {
	doc := dom.Doc()
	doc.Body().SetHTML("<div id='root'><span>old</span></div>")

	BindEl(doc.ByID("root"), func(el El) {
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

func TestGroupForEach(t *testing.T) {
	cards := NewGroup()
	Div().Text("a").Group(cards)
	Div().Text("b").Group(cards)

	var texts []string
	cards.ForEach(func(el dom.Element) {
		texts = append(texts, el.Text())
	})

	if len(texts) != 2 || texts[0] != "a" || texts[1] != "b" {
		t.Fatalf("unexpected texts %v", texts)
	}
}

func TestGroupMerge(t *testing.T) {
	g1 := Group(Div().Text("a"))
	g2 := Group(Div().Text("b"))
	g1.Group(g2)

	var texts []string
	g1.ForEach(func(el dom.Element) {
		texts = append(texts, el.Text())
	})

	if len(texts) != 2 || texts[0] != "a" || texts[1] != "b" {
		t.Fatalf("unexpected texts %v", texts)
	}
}

func TestGroupInvalidArgs(t *testing.T) {
	assertPanics(t, func() { Group() })
	cards := Group(Div())
	assertPanics(t, func() { cards.ForEach(nil) })
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

func TestGroupMutators(t *testing.T) {
	g := Group(Div(), Div())

	g.SetText("hi")
	g.ForEach(func(el dom.Element) {
		if el.Text() != "hi" {
			t.Fatalf("expected text hi")
		}
	})

	g.SetHTML("<span>ok</span>")
	g.AddClass("x").ToggleClass("y").RemoveClass("x")
	g.SetAttr("data-x", "1").SetStyle("color", "red")

	g.ForEach(func(el dom.Element) {
		if el.HTML() != "<span>ok</span>" {
			t.Fatalf("unexpected html %q", el.HTML())
		}
		if el.HasClass("x") {
			t.Fatalf("class x should be removed")
		}
		if !el.HasClass("y") {
			t.Fatalf("expected class y")
		}
		if v := el.Attr("data-x"); v != "1" {
			t.Fatalf("expected attr data-x=1, got %q", v)
		}
		if v := el.Get("style").Call("getPropertyValue", "color").String(); v != "red" {
			t.Fatalf("expected style color red, got %q", v)
		}
	})
}

func TestAnchorBuilder(t *testing.T) {
	a := A().Class("c").Href("/docs").Text("docs")
	el := a.Element()
	if v := el.Attr("href"); v != "/docs" {
		t.Fatalf("expected href /docs, got %q", v)
	}
	if el.Text() != "docs" {
		t.Fatalf("expected text docs")
	}
	if !el.HasClass("c") {
		t.Fatalf("expected class c")
	}
}
