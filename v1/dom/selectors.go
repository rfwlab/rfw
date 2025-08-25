//go:build js && wasm

package dom

import (
	jst "syscall/js"

	js "github.com/rfwlab/rfw/v1/js"
)

// Element represents a DOM element returned by selector helpers.
type Element = jst.Value

// CreateElement returns a new element with the given tag name.
func CreateElement(tag string) Element {
	return js.Doc().Call("createElement", tag)
}

// ByID fetches an element by its id attribute.
func ByID(id string) Element {
	return js.Doc().Call("getElementById", id)
}

// Query returns the first element matching the CSS selector.
func Query(selector string) Element {
	return js.Doc().Call("querySelector", selector)
}

// QueryAll returns all elements matching the CSS selector.
func QueryAll(selector string) Element {
	return js.Doc().Call("querySelectorAll", selector)
}

// ByClass returns all elements with the given class name.
func ByClass(name string) Element {
	return js.Doc().Call("getElementsByClassName", name)
}

// ByTag returns all elements with the given tag name.
func ByTag(tag string) Element {
	return js.Doc().Call("getElementsByTagName", tag)
}

// SetInnerHTML replaces an element's children with the provided HTML string.
func SetInnerHTML(el Element, html string) {
	el.Set("innerHTML", html)
}

// Text returns an element's text content.
func Text(el Element) string {
	return el.Get("textContent").String()
}

// SetText sets an element's text content.
func SetText(el Element, text string) {
	el.Set("textContent", text)
}

// Attr retrieves the value of an attribute or an empty string if unset.
func Attr(el Element, name string) string {
	v := el.Call("getAttribute", name)
	if v.Truthy() {
		return v.String()
	}
	return ""
}

// SetAttr sets the value of an attribute on the element.
func SetAttr(el Element, name, value string) {
	el.Call("setAttribute", name, value)
}

// AddClass adds a class to the element's class list.
func AddClass(el Element, class string) {
	el.Get("classList").Call("add", class)
}

// RemoveClass removes a class from the element's class list.
func RemoveClass(el Element, class string) {
	el.Get("classList").Call("remove", class)
}
