//go:build js && wasm

package composition

import "github.com/rfwlab/rfw/v1/dom"

// Node represents a DOM node that can be appended to other nodes.
type Node interface {
	Element() dom.Element
}

// El wraps a DOM element exposed to Bind and For callbacks.
type El interface {
	Clear()
	Append(nodes ...Node)
}

type elWrap struct{ dom.Element }

func (e elWrap) Clear() { e.SetHTML("") }

func (e elWrap) Append(nodes ...Node) {
	for _, n := range nodes {
		if n != nil {
			e.AppendChild(n.Element())
		}
	}
}

// Elements groups a collection of DOM elements for bulk operations.
type Elements struct{ els []dom.Element }

// NewGroup creates an empty Elements collection.
func NewGroup() *Elements { return &Elements{} }

// Group collects provided nodes into an Elements wrapper without relying on selectors.
func Group(nodes ...Node) *Elements {
	if len(nodes) == 0 {
		panic("composition.Group: no nodes")
	}
	g := NewGroup()
	g.add(nodes...)
	return g
}

// Group merges the current group with other groups.
func (g *Elements) Group(gs ...*Elements) *Elements {
	for _, other := range gs {
		if other != nil {
			g.els = append(g.els, other.els...)
		}
	}
	return g
}

// ForEach invokes fn for each element in the group.
func (g *Elements) ForEach(fn func(dom.Element)) {
	if fn == nil {
		panic("composition.Elements.ForEach: nil fn")
	}
	for _, el := range g.els {
		fn(el)
	}
}

// AddClass adds a class to every element in the group.
func (g *Elements) AddClass(name string) *Elements {
	for _, el := range g.els {
		el.AddClass(name)
	}
	return g
}

// RemoveClass removes a class from every element in the group.
func (g *Elements) RemoveClass(name string) *Elements {
	for _, el := range g.els {
		el.RemoveClass(name)
	}
	return g
}

// ToggleClass toggles a class on every element in the group.
func (g *Elements) ToggleClass(name string) *Elements {
	for _, el := range g.els {
		el.ToggleClass(name)
	}
	return g
}

// SetAttr sets an attribute on every element in the group.
func (g *Elements) SetAttr(name, value string) *Elements {
	for _, el := range g.els {
		el.SetAttr(name, value)
	}
	return g
}

// SetStyle sets an inline style property on every element in the group.
func (g *Elements) SetStyle(prop, value string) *Elements {
	for _, el := range g.els {
		el.SetStyle(prop, value)
	}
	return g
}

// SetText sets the text content of every element in the group.
func (g *Elements) SetText(t string) *Elements {
	for _, el := range g.els {
		el.SetText(t)
	}
	return g
}

// SetHTML replaces the HTML of every element in the group.
func (g *Elements) SetHTML(html string) *Elements {
	for _, el := range g.els {
		el.SetHTML(html)
	}
	return g
}

func (g *Elements) add(nodes ...Node) {
	for _, n := range nodes {
		if n != nil {
			g.els = append(g.els, n.Element())
		}
	}
}

// BindEl invokes fn with a wrapper exposing Clear and Append helpers for the
// provided element.
func BindEl(el dom.Element, fn func(El)) {
	if fn == nil {
		panic("composition.BindEl: nil fn")
	}
	if el.IsNull() || el.IsUndefined() {
		return
	}
	fn(elWrap{el})
}

// Bind selects the first element matching selector and invokes fn with a
// wrapper exposing Clear and Append helpers.
func Bind(selector string, fn func(El)) {
	if selector == "" {
		panic("composition.Bind: empty selector")
	}
	if fn == nil {
		panic("composition.Bind: nil fn")
	}
	el := dom.Doc().Query(selector)
	BindEl(el, fn)
}

// For repeatedly calls fn to generate nodes and appends them to the element
// matched by selector via Bind. Iteration stops when fn returns nil.
func For(selector string, fn func() Node) {
	if fn == nil {
		panic("composition.For: nil fn")
	}
	Bind(selector, func(e El) {
		for {
			n := fn()
			if n == nil {
				break
			}
			e.Append(n)
		}
	})
}

// divNode builds a <div> element.
type divNode struct{ el dom.Element }

// Div creates a new <div> node builder.
func Div() *divNode { return &divNode{el: dom.Doc().CreateElement("div")} }

// Element returns the underlying DOM element.
func (d *divNode) Element() dom.Element { return d.el }

// Class adds a class to the element.
func (d *divNode) Class(name string) *divNode {
	d.el.AddClass(name)
	return d
}

// Classes adds multiple classes to the element.
func (d *divNode) Classes(names ...string) *divNode {
	for _, name := range names {
		d.el.AddClass(name)
	}
	return d
}

// Style sets an inline style property on the element.
func (d *divNode) Style(prop, value string) *divNode {
	d.el.SetStyle(prop, value)
	return d
}

// Styles adds multiple inline style properties to the element.
func (d *divNode) Styles(props ...string) *divNode {
	for i := 0; i < len(props); i += 2 {
		d.el.SetStyle(props[i], props[i+1])
	}
	return d
}

// Text sets the text content of the element.
func (d *divNode) Text(t string) *divNode {
	d.el.SetText(t)
	return d
}

// Group adds the node to the provided group.
func (d *divNode) Group(g *Elements) *divNode {
	if g != nil {
		g.add(d)
	}
	return d
}

// anchorNode builds an <a> element.
type anchorNode struct{ el dom.Element }

// A creates a new <a> node builder.
func A() *anchorNode { return &anchorNode{el: dom.Doc().CreateElement("a")} }

// Element returns the underlying DOM element.
func (a *anchorNode) Element() dom.Element { return a.el }

// Class adds a class to the element.
func (a *anchorNode) Class(name string) *anchorNode {
	a.el.AddClass(name)
	return a
}

// Classes adds multiple classes to the element.
func (a *anchorNode) Classes(names ...string) *anchorNode {
	for _, name := range names {
		a.el.AddClass(name)
	}
	return a
}

// Style sets an inline style property on the element.
func (a *anchorNode) Style(prop, value string) *anchorNode {
	a.el.SetStyle(prop, value)
	return a
}

// Styles adds multiple inline style properties to the element.
func (a *anchorNode) Styles(props ...string) *anchorNode {
	for i := 0; i < len(props); i += 2 {
		a.el.SetStyle(props[i], props[i+1])
	}
	return a
}

// Attr sets an attribute on the element.
func (a *anchorNode) Attr(name, value string) *anchorNode {
	a.el.SetAttr(name, value)
	return a
}

// Href sets the href attribute on the element.
func (a *anchorNode) Href(h string) *anchorNode {
	a.el.SetAttr("href", h)
	return a
}

// Text sets the text content of the element.
func (a *anchorNode) Text(t string) *anchorNode {
	a.el.SetText(t)
	return a
}

// Group adds the node to the provided group.
func (a *anchorNode) Group(g *Elements) *anchorNode {
	if g != nil {
		g.add(a)
	}
	return a
}
