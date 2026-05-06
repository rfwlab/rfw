# dom

```go
import "github.com/rfwlab/rfw/v2/dom"
```

Low-level DOM helpers. Most apps interact with DOM indirectly via components.

## Accessors

| Function | Description |
| --- | --- |
| `Doc() *Document` | Returns the global document. |
| `Window() *Window` | Returns the global window. |
| `Document.Head() *Element` | Returns `<head>`. |
| `Document.Body() *Element` | Returns `<body>`. |

## Queries

| Function | Description |
| --- | --- |
| `ByID(id string) *Element` | Fetch element by id. |
| `Query(sel string) *Element` | First element matching selector. |
| `QueryAll(sel string) []*Element` | All elements matching selector. |
| `ByClass(name string) []*Element` | Elements with given class. |
| `ByTag(tag string) []*Element` | Elements with given tag. |

## Component Root

| Function | Description |
| --- | --- |
| `ComponentRoot(id string) Element` | Returns the DOM root element for a component by its ID. Falls back to `#app` if id is empty or element not found. |

`ComponentRoot` is the standard way to look up a component's root element. It replaces the `doc.Query(fmt.Sprintf("[data-component-id='%s']", id))` pattern used throughout the framework.

## Events

| Function | Description |
| --- | --- |
| `DelegateEvents(root *Element)` | Set up event delegation from root. |
| `RegisterHandlerFunc(name string, fn func(*Event))` | Register a named handler. |
| `GetHandler(name string) func(*Event)` | Retrieve a named handler. |

## Mutation

| Function | Description |
| --- | --- |
| `UpdateDOM(id, html string)` | Patch a component's DOM with raw HTML. |
| `CreateElement(tag string) *Element` | Create a new element. |

## Element Methods

| Method | Description |
| --- | --- |
| `SetHTML(html string)` | Replace children with raw HTML. |
| `SetText(text string)` | Set text content. |
| `AddClass(name string)` | Add a CSS class. |
| `RemoveClass(name string)` | Remove a CSS class. |
| `HasClass(name string) bool` | Check for a CSS class. |
| `ToggleClass(name string)` | Toggle a CSS class. |
| `SetAttr(name, value string)` | Set an attribute. |
| `Attr(name string) string` | Get attribute value. |
| `SetStyle(prop, value string)` | Set inline style property. |
| `AppendChild(child *Element)` | Append a child element. |
| `On(event string, handler func(*Event)) func()` | Attach listener; returns cleanup. |