//go:build js && wasm

package dom

import js "github.com/rfwlab/rfw/v1/js"

// Element wraps a DOM element and provides typed helpers.
type Element struct{ js.Value }

// Query returns the first descendant matching the CSS selector.
func (e Element) Query(sel string) Element {
	return Element{e.Call("querySelector", sel)}
}

// QueryAll returns all descendants matching the selector.
func (e Element) QueryAll(sel string) Element {
	return Element{e.Call("querySelectorAll", sel)}
}

// ByClass returns all descendants with the given class name.
func (e Element) ByClass(name string) Element {
	return Element{e.Call("getElementsByClassName", name)}
}

// ByTag returns all descendants with the given tag name.
func (e Element) ByTag(tag string) Element {
	return Element{e.Call("getElementsByTagName", tag)}
}

// Text returns the element's text content.
func (e Element) Text() string { return e.Get("textContent").String() }

// SetText sets the element's text content.
func (e Element) SetText(txt string) { e.Set("textContent", txt) }

// HTML returns the element's inner HTML.
func (e Element) HTML() string { return e.Get("innerHTML").String() }

// SetHTML replaces the element's children with raw HTML.
func (e Element) SetHTML(html string) { e.Set("innerHTML", html) }

// Attr retrieves the value of an attribute or "" if unset.
func (e Element) Attr(name string) string {
	v := e.Call("getAttribute", name)
	if v.Truthy() {
		return v.String()
	}
	return ""
}

// SetAttr sets the value of an attribute on the element.
func (e Element) SetAttr(name, value string) { e.Call("setAttribute", name, value) }

// SetStyle sets an inline style property on the element.
func (e Element) SetStyle(prop, value string) {
	e.Get("style").Call("setProperty", prop, value)
}

// AddClass adds a class to the element.
func (e Element) AddClass(name string) { e.Get("classList").Call("add", name) }

// RemoveClass removes a class from the element.
func (e Element) RemoveClass(name string) { e.Get("classList").Call("remove", name) }

// HasClass reports whether the element has the given class.
func (e Element) HasClass(name string) bool {
	return e.Get("classList").Call("contains", name).Bool()
}

// ToggleClass toggles the presence of a class on the element.
func (e Element) ToggleClass(name string) {
	e.Get("classList").Call("toggle", name)
}

// Length returns the number of children when the element represents a collection.
func (e Element) Length() int { return e.Get("length").Int() }

// Index retrieves the element at the given position when representing a collection.
func (e Element) Index(i int) Element { return Element{e.Value.Index(i)} }
