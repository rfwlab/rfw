# dom

Low level DOM helpers used by the framework. Most applications interact
with the DOM indirectly through components, but the following utilities
are available for advanced use:

| Function | Description |
| --- | --- |
| `CreateElement(tag)` | Returns a new element. |
| `ByID(id)` | Fetches an element by id. |
| `Query(sel)` | Returns the first element matching the CSS selector. |
| `QueryAll(sel)` | Returns all elements matching the CSS selector. |
| `ByClass(name)` | Returns elements with the given class name. |
| `ByTag(tag)` | Returns elements with the given tag name. |
| `SetInnerHTML(el, html)` | Replaces an element's children with raw HTML. |
| `Text(el)` | Returns an element's text content. |
| `SetText(el, text)` | Sets an element's text content. |
| `Attr(el, name)` | Retrieves the value of an attribute. |
| `SetAttr(el, name, value)` | Sets the value of an attribute. |
| `AddClass(el, name)` | Adds a class to an element. |
| `RemoveClass(el, name)` | Removes a class from an element. |

## Usage

Besides the methods shown, the package exposes `RegisterHandlerFunc` to bind
Go functions to named DOM events.

> **Note**
> Prefer these helpers over the low-level `js` package. They shield your code
> from browser globals, making components more portable and easier to test.
> If a helper is missing, fall back to a `js` alias but avoid `js.Global`.

### Selecting and modifying elements

```go
title := dom.Query("h1")
dom.SetText(title, "Hello")
dom.AddClass(title, "highlight")
```

The snippet demonstrates direct DOM interactions.

@include:ExampleFrame:{code:"/examples/components/event_component.go", uri:"/examples/event"}
