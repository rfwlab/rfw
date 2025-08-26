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
| `ScheduleRender(id, html, delay)` | Updates a component after a delay. |

## Usage

Besides the methods shown, the package exposes `RegisterHandlerFunc` to bind
Go functions to named DOM events.

> **Note**
> Prefer these helpers over the low-level `js` package, which centralizes
> wrappers around `syscall/js`. They shield your code from browser globals,
> making components more portable and easier to test. If a helper is missing,
> fall back to the `js` package but avoid `js.Global`.

### Selecting and modifying elements

```go
title := dom.Query("h1")
dom.SetText(title, "Hello")
dom.AddClass(title, "highlight")
```

The snippet demonstrates direct DOM interactions.

@include:ExampleFrame:{code:"/examples/components/event_component.go", uri:"/examples/event"}

### Delayed updates

```go
dom.ScheduleRender("comp", "<p>later</p>", time.Second)
```

### Virtual lists

The `virtual` subpackage exposes a `<VirtualList>` component that only renders
items visible within a scroll container.

```go
import "github.com/rfwlab/rfw/v1/dom/virtual"

list := virtual.NewVirtualList("list", 1000, 24, func(i int) string {
    return fmt.Sprintf("<div class='item'>Row %d</div>", i)
})
```

```html
<div id="list" style="height:200px; overflow-y:auto"></div>
```
