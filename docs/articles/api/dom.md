# dom

Low level DOM helpers used by the framework. Most applications interact
with the DOM indirectly through components, but the following utilities
are available for advanced use:

| Function | Description |
| --- | --- |
| `CreateElement(tag)` | Returns a new element. |
| `ByID(id)` | Fetches an element by id. |
| `SetInnerHTML(el, html)` | Replaces an element's children with raw HTML. |

## Usage

Besides the methods shown, the package exposes `RegisterHandlerFunc` to bind
Go functions to named DOM events.

## Example

```go
dom.RegisterHandlerFunc("increment", func() {
        if val, ok := c.Store.Get("count").(int); ok {
                c.Store.Set("count", val+1)
        }
})
```

1. `dom.RegisterHandlerFunc` associates the `increment` identifier with a
   function.
2. When called from the DOM, the function reads `count` from the store and
   increments it.
