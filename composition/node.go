//go:build js && wasm

package composition

import (
	"fmt"

	"github.com/rfwlab/rfw/v2/dom"
)

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

// DivNode builds a <div> element.
type DivNode struct{ el dom.Element }

// Div creates a new <div> node builder.
func Div() *DivNode { return &DivNode{el: dom.Doc().CreateElement("div")} }

// Element returns the underlying DOM element.
func (d *DivNode) Element() dom.Element { return d.el }

// Class adds a class to the element.
func (d *DivNode) Class(name string) *DivNode {
	d.el.AddClass(name)
	return d
}

// Classes adds multiple classes to the element.
func (d *DivNode) Classes(names ...string) *DivNode {
	for _, name := range names {
		d.el.AddClass(name)
	}
	return d
}

// Style sets an inline style property on the element.
func (d *DivNode) Style(prop, value string) *DivNode {
	d.el.SetStyle(prop, value)
	return d
}

// Styles adds multiple inline style properties to the element.
func (d *DivNode) Styles(props ...string) *DivNode {
	for i := 0; i < len(props); i += 2 {
		d.el.SetStyle(props[i], props[i+1])
	}
	return d
}

// Text sets the text content of the element.
func (d *DivNode) Text(t string) *DivNode {
	d.el.SetText(t)
	return d
}

// Group adds the node to the provided group.
func (d *DivNode) Group(g *Elements) *DivNode {
	if g != nil {
		g.add(d)
	}
	return d
}

// AnchorNode builds an <a> element.
type AnchorNode struct{ el dom.Element }

// A creates a new <a> node builder.
func A() *AnchorNode { return &AnchorNode{el: dom.Doc().CreateElement("a")} }

// Element returns the underlying DOM element.
func (a *AnchorNode) Element() dom.Element { return a.el }

// Class adds a class to the element.
func (a *AnchorNode) Class(name string) *AnchorNode {
	a.el.AddClass(name)
	return a
}

// Classes adds multiple classes to the element.
func (a *AnchorNode) Classes(names ...string) *AnchorNode {
	for _, name := range names {
		a.el.AddClass(name)
	}
	return a
}

// Style sets an inline style property on the element.
func (a *AnchorNode) Style(prop, value string) *AnchorNode {
	a.el.SetStyle(prop, value)
	return a
}

// Styles adds multiple inline style properties to the element.
func (a *AnchorNode) Styles(props ...string) *AnchorNode {
	for i := 0; i < len(props); i += 2 {
		a.el.SetStyle(props[i], props[i+1])
	}
	return a
}

// Attr sets an attribute on the element.
func (a *AnchorNode) Attr(name, value string) *AnchorNode {
	a.el.SetAttr(name, value)
	return a
}

// Href sets the href attribute on the element.
func (a *AnchorNode) Href(h string) *AnchorNode {
	a.el.SetAttr("href", h)
	return a
}

// Text sets the text content of the element.
func (a *AnchorNode) Text(t string) *AnchorNode {
	a.el.SetText(t)
	return a
}

// Group adds the node to the provided group.
func (a *AnchorNode) Group(g *Elements) *AnchorNode {
	if g != nil {
		g.add(a)
	}
	return a
}

// SpanNode builds a <span> element.
type SpanNode struct{ el dom.Element }

// Span creates a new <span> node builder.
func Span() *SpanNode { return &SpanNode{el: dom.Doc().CreateElement("span")} }

// Element returns the underlying DOM element.
func (s *SpanNode) Element() dom.Element { return s.el }

// Class adds a class to the element.
func (s *SpanNode) Class(name string) *SpanNode {
	s.el.AddClass(name)
	return s
}

// Classes adds multiple classes to the element.
func (s *SpanNode) Classes(names ...string) *SpanNode {
	for _, name := range names {
		s.el.AddClass(name)
	}
	return s
}

// Style sets an inline style property on the element.
func (s *SpanNode) Style(prop, value string) *SpanNode {
	s.el.SetStyle(prop, value)
	return s
}

// Styles adds multiple inline style properties to the element.
func (s *SpanNode) Styles(props ...string) *SpanNode {
	for i := 0; i < len(props); i += 2 {
		s.el.SetStyle(props[i], props[i+1])
	}
	return s
}

// Text sets the text content of the element.
func (s *SpanNode) Text(t string) *SpanNode {
	s.el.SetText(t)
	return s
}

// Group adds the node to the provided group.
func (s *SpanNode) Group(g *Elements) *SpanNode {
	if g != nil {
		g.add(s)
	}
	return s
}

// ButtonNode builds a <button> element.
type ButtonNode struct{ el dom.Element }

// Button creates a new <button> node builder.
func Button() *ButtonNode { return &ButtonNode{el: dom.Doc().CreateElement("button")} }

// Element returns the underlying DOM element.
func (b *ButtonNode) Element() dom.Element { return b.el }

// Class adds a class to the element.
func (b *ButtonNode) Class(name string) *ButtonNode {
	b.el.AddClass(name)
	return b
}

// Classes adds multiple classes to the element.
func (b *ButtonNode) Classes(names ...string) *ButtonNode {
	for _, name := range names {
		b.el.AddClass(name)
	}
	return b
}

// Style sets an inline style property on the element.
func (b *ButtonNode) Style(prop, value string) *ButtonNode {
	b.el.SetStyle(prop, value)
	return b
}

// Styles adds multiple inline style properties to the element.
func (b *ButtonNode) Styles(props ...string) *ButtonNode {
	for i := 0; i < len(props); i += 2 {
		b.el.SetStyle(props[i], props[i+1])
	}
	return b
}

// Text sets the text content of the element.
func (b *ButtonNode) Text(t string) *ButtonNode {
	b.el.SetText(t)
	return b
}

// Group adds the node to the provided group.
func (b *ButtonNode) Group(g *Elements) *ButtonNode {
	if g != nil {
		g.add(b)
	}
	return b
}

// HeadingNode builds an <h1>..<h6> element.
type HeadingNode struct{ el dom.Element }

// H creates a new heading node builder for level 1..6 (out-of-range coerced to 1..6).
func H(level int) *HeadingNode {
	if level < 1 {
		level = 1
	} else if level > 6 {
		level = 6
	}
	tag := fmt.Sprintf("h%d", level)
	return &HeadingNode{el: dom.Doc().CreateElement(tag)}
}

// Element returns the underlying DOM element.
func (h *HeadingNode) Element() dom.Element { return h.el }

// Class adds a class to the element.
func (h *HeadingNode) Class(name string) *HeadingNode {
	h.el.AddClass(name)
	return h
}

// Classes adds multiple classes to the element.
func (h *HeadingNode) Classes(names ...string) *HeadingNode {
	for _, name := range names {
		h.el.AddClass(name)
	}
	return h
}

// Style sets an inline style property on the element.
func (h *HeadingNode) Style(prop, value string) *HeadingNode {
	h.el.SetStyle(prop, value)
	return h
}

// Styles adds multiple inline style properties to the element.
func (h *HeadingNode) Styles(props ...string) *HeadingNode {
	for i := 0; i < len(props); i += 2 {
		h.el.SetStyle(props[i], props[i+1])
	}
	return h
}

// Text sets the text content of the element.
func (h *HeadingNode) Text(t string) *HeadingNode {
	h.el.SetText(t)
	return h
}

// Group adds the node to the provided group.
func (h *HeadingNode) Group(g *Elements) *HeadingNode {
	if g != nil {
		g.add(h)
	}
	return h
}
