# dom

Low level DOM helpers used by the framework. Most applications interact
with the DOM indirectly through components.

## Context

These helpers allow fine-grained control over the browser DOM without
relying on `syscall/js` directly.

## Prerequisites

Use these APIs when components need to manipulate elements that are not
handled by the template system.

## API

Both function helpers and typed wrappers are available. The functions
delegate to the wrappers for backward compatibility:

| Function | Description |
| --- | --- |
| `Doc()` | Returns the global `Document`. |
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
| `SetStyle(el, prop, value)` | Sets an inline style property. |
| `StyleInline(map)` | Builds a style string from CSS properties. |
| `AddClass(el, name)` | Adds a class to an element. |
| `RemoveClass(el, name)` | Removes a class from an element. |
| `HasClass(el, name)` | Reports whether an element has a class. |
| `ToggleClass(el, name)` | Toggles a class on an element. |
| `ScheduleRender(id, html, delay)` | Updates a component after a delay. |
| `UpdateDOM(id, html)` | Patches a component's DOM with raw HTML. |
| `TemplateHook` | Callback invoked after `UpdateDOM`. |
| `BindStoreInputs(el)` | Binds inputs to `@store` directives. |
| `BindSignalInputs(id, el)` | Binds inputs to local `@signal` directives. |

## Usage

Besides the methods shown, the package exposes `RegisterHandlerFunc` to bind
Go functions to named DOM events.

> **Note**
> Prefer these helpers over the low-level `js` package, which centralizes
> wrappers around `syscall/js`. They shield your code from browser globals,
> making components more portable and easier to test. If a helper is missing,
> fall back to the `js` package but avoid `js.Global`.

### Selecting and modifying elements

1. Obtain the document wrapper.
   ```go
   doc := dom.Doc()
   ```
2. Query and update elements with typed methods.
   ```go
   title := doc.Query("h1")
   title.SetText("Hello")
   title.AddClass("highlight")
   title.SetAttr("lang", "en")
   title.SetStyle("color", "red")

   box := doc.CreateElement("div")
   box.SetAttr("style", dom.StyleInline(map[string]string{"display": "flex", "gap": "4px"}))

   items := doc.QueryAll("li")
   for i := 0; i < items.Length(); i++ {
       items.Index(i).ToggleClass("active")
   }
   ```

The snippet demonstrates direct DOM interactions.

`QueryAll` returns a collection exposing `Length` and `Index` to walk matched elements.

### Migration

Legacy helpers such as `dom.Query` and `dom.SetText` continue to work and
delegate to the new methods for compatibility.

@include:ExampleFrame:{code:"/examples/components/event_component.go", uri:"/examples/event"}

### Delayed updates

```go
dom.ScheduleRender("comp", "<p>later</p>", time.Second)
```

### Custom rendering and manual bindings

#### Manual DOM updates

`UpdateDOM` applies raw HTML to a component and re-binds event handlers.
Assign `TemplateHook` to inspect or post-process the generated markup:

```go
dom.TemplateHook = func(id, html string) {
    log.Printf("patched %s", id)
}

dom.UpdateDOM("profile", "<p>signed in</p>")
```

#### Manual input binding

`UpdateDOM` automatically connects inputs with `@store` and `@signal`
directives. When building elements yourself, call the bind helpers
explicitly:

```go
import (
    "github.com/rfwlab/rfw/v1/dom"
    "github.com/rfwlab/rfw/v1/state"
)

doc := dom.Doc()
el := doc.CreateElement("div")
el.SetHTML(`<input value="@store:user.name:w">`)
dom.BindStoreInputs(el.Value)

nameSig := state.NewSignal("")
dom.RegisterSignal("cmp", "name", nameSig)
el.SetHTML(`<input value="@signal:name:w">`)
dom.BindSignalInputs("cmp", el.Value)
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

### Notes and limitations

Only common operations are wrapped. For unsupported features, use the
`dom` helpers or the lower level `js` package cautiously.

### Related links

- [js](js.md)
