# dom

Low level DOM helpers used by the framework. Most applications interact with the DOM indirectly through components.

| Function | Description |
| --- | --- |
| `Doc()` | Returns the global `Document`. |
| `Document.Head()` | Returns the `<head>` element. |
| `Document.Body()` | Returns the `<body>` element. |
| `CreateElement(tag)` | Returns a new element. |
| `ByID(id)` | Fetches an element by id. |
| `Query(sel)` | Returns the first element matching the CSS selector. |
| `QueryAll(sel)` | Returns all elements matching the CSS selector. |
| `ByClass(name)` | Returns elements with the given class name. |
| `ByTag(tag)` | Returns elements with the given tag name. |
| `SetInnerHTML(el, html)` | Replaces an element's children with raw HTML. |
| `Element.AppendChild(child)` | Appends a child to an element. |
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
| `(Element).On(event, handler)` | Attaches an event listener and returns a stop function. |
| `(Element).OnClick(handler)` | Convenience wrapper for click events. |

