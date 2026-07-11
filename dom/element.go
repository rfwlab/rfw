//go:build js && wasm

package dom

import js "github.com/rfwlab/rfw/v2/js"

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

// AppendChild appends a child element.
func (e Element) AppendChild(child Element) { e.Call("appendChild", child.Value) }

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

// Val returns the element's value property (inputs, selects, textareas).
// Named Val because the embedded js.Value field occupies Value.
func (e Element) Val() string { return e.Get("value").String() }

// SetValue sets the element's value property.
func (e Element) SetValue(v string) { e.Set("value", v) }

// Checked reports whether a checkbox or radio input is checked.
func (e Element) Checked() bool { return e.Get("checked").Bool() }

// Data reads a data-* attribute by its dataset key (camelCase: data-item-id
// becomes Data("itemId")).
func (e Element) Data(key string) string {
	v := e.Get("dataset").Get(key)
	if !v.Truthy() {
		return ""
	}
	return v.String()
}

// Closest returns the nearest ancestor (or the element itself) matching the
// selector; check IsNull on the result for no match.
func (e Element) Closest(sel string) Element {
	return Element{e.Call("closest", sel)}
}
