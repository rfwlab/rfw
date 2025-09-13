//go:build js && wasm

package dom

import js "github.com/rfwlab/rfw/v1/js"

// Document wraps the global document object.
type Document struct{ js.Value }

// Doc returns the global Document.
func Doc() Document { return Document{js.Doc()} }

// ByID fetches an element by id.
func (d Document) ByID(id string) Element {
	return Element{d.Call("getElementById", id)}
}

// Query returns the first element matching the selector.
func (d Document) Query(sel string) Element {
	return Element{d.Call("querySelector", sel)}
}

// QueryAll returns all elements matching the selector.
func (d Document) QueryAll(sel string) Element {
	return Element{d.Call("querySelectorAll", sel)}
}

// ByClass returns all elements with the given class name.
func (d Document) ByClass(name string) Element {
	return Element{d.Call("getElementsByClassName", name)}
}

// ByTag returns all elements with the given tag name.
func (d Document) ByTag(tag string) Element {
	return Element{d.Call("getElementsByTagName", tag)}
}

// CreateElement creates a new element with the tag.
func (d Document) CreateElement(tag string) Element {
	return Element{d.Call("createElement", tag)}
}

// Head returns the document's <head> element.
func (d Document) Head() Element { return Element{d.Get("head")} }

// Body returns the document's <body> element.
func (d Document) Body() Element { return Element{d.Get("body")} }
