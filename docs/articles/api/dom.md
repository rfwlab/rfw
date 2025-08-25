# dom

Low level DOM helpers used by the framework. Most applications interact
with the DOM indirectly through components, but the following utilities
are available for advanced use:

| Function | Description |
| --- | --- |
| `CreateElement(tag)` | Returns a new element. |
| `ByID(id)` | Fetches an element by id. |
| `SetInnerHTML(el, html)` | Replaces an element's children with raw HTML. |
| `QueryAll(sel)` | Returns all elements matching the CSS selector. |

## Usage

Besides the methods shown, the package exposes `RegisterHandlerFunc` to bind
Go functions to named DOM events.

> **Note**
> Prefer these helpers over the low-level `js` package. They shield your code
> from browser globals, making components more portable and easier to test.
> If a helper is missing, fall back to a `js` alias but avoid `js.Global`.

The snippet demonstrates direct DOM interactions.

@include:ExampleFrame:{code:"/examples/components/event_component.go", uri:"/examples/event"}
