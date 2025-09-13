//go:build js && wasm

package dom

// CreateElement returns a new element with the given tag name.
func CreateElement(tag string) Element { return Doc().CreateElement(tag) }

// ByID fetches an element by its id attribute.
func ByID(id string) Element { return Doc().ByID(id) }

// Query returns the first element matching the CSS selector.
func Query(selector string) Element { return Doc().Query(selector) }

// QueryAll returns all elements matching the CSS selector.
func QueryAll(selector string) Element { return Doc().QueryAll(selector) }

// ByClass returns all elements with the given class name.
func ByClass(name string) Element { return Doc().ByClass(name) }

// ByTag returns all elements with the given tag name.
func ByTag(tag string) Element { return Doc().ByTag(tag) }

// SetInnerHTML replaces an element's children with the provided HTML string.
func SetInnerHTML(el Element, html string) { el.SetHTML(html) }

// Text returns an element's text content.
func Text(el Element) string { return el.Text() }

// SetText sets an element's text content.
func SetText(el Element, text string) { el.SetText(text) }

// Attr retrieves the value of an attribute or an empty string if unset.
func Attr(el Element, name string) string { return el.Attr(name) }

// SetAttr sets the value of an attribute on the element.
func SetAttr(el Element, name, value string) { el.SetAttr(name, value) }

// AddClass adds a class to the element's class list.
func AddClass(el Element, class string) { el.AddClass(class) }

// RemoveClass removes a class from the element's class list.
func RemoveClass(el Element, class string) { el.RemoveClass(class) }

// HasClass reports whether the element has the specified class.
func HasClass(el Element, class string) bool { return el.HasClass(class) }

// ToggleClass toggles the presence of a class on the element's class list.
func ToggleClass(el Element, class string) { el.ToggleClass(class) }

// SetStyle sets an inline style property on the element.
func SetStyle(el Element, prop, value string) { el.SetStyle(prop, value) }
