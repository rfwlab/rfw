# dom

Low level DOM helpers used by the framework. Most applications interact
with the DOM indirectly through components, but the following utilities
are available for advanced use:

| Function | Description |
| --- | --- |
| `CreateElement(tag)` | Returns a new element. |
| `ByID(id)` | Fetches an element by id. |
| `SetInnerHTML(el, html)` | Replaces an element's children with raw HTML. |
