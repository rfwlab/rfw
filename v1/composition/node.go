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
	if el.IsNull() || el.IsUndefined() {
		return
	}
	fn(elWrap{el})
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
