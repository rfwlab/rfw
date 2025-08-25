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

// QueryAll returns all elements matching the CSS selector.
func QueryAll(selector string) Element {
	return js.Doc().Call("querySelectorAll", selector)
}

// SetInnerHTML replaces an element's children with the provided HTML string.
func SetInnerHTML(el Element, html string) {
	el.Set("innerHTML", html)
}
