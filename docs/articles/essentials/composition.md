# Composition

The `composition` package is the core of rfw v2's component system. `composition.New` replaces v1's manual wiring with tag-driven auto-composition.

---

## How `composition.New` Works

`composition.New(&MyStruct{})` performs these steps:

1. **Scan struct tags**, Uses the `scan` package to parse all `rfw:` tags on the struct fields.
2. **Resolve template**, Checks for `rfw:"template:path"`, then falls back to convention (struct name → `Name.rtml`).
3. **Create HTML component**, Calls `core.NewHTMLComponent` with the resolved template.
4. **Wrap as Component**, `Wrap(hc)` creates a `composition.Component` providing `Prop`, `On`, `Store`, and `History` methods.
5. **Initialize default store**, Creates or reuses the `"default"` store under the `"app"` module.
6. **Wire signals**, For each `rfw:"signal"` field: if non-nil, registers it as a prop; if nil, creates a zero-value signal, sets the field, and registers as prop.
7. **Wire props**, Creates nil `Signal[any]` props for any `rfw:"prop"` fields not already in the component's props map.
8. **Register event handlers**, For `rfw:"event:domEvent:handler"` tags, looks up the method by name and calls `comp.On(handler, fn)`.
9. **Auto-discover zero-arg methods**, Any exported no-argument, no-return method on the struct (excluding `Component` methods) is auto-registered as a handler under its method name.
10. **Wire lifecycle**, If `OnMount()` or `OnUnmount()` exist as no-arg methods, they are wired into the component's mount/unmount callbacks.
11. **Create stores**, For each `rfw:"store"` tag, calls `comp.Store(name)`.
12. **Register hosts**, For `rfw:"host"` tags, calls `AddHostComponent(name)`.
13. **Wire histories**, For `rfw:"history:store:undo:redo"` tags, retrieves the store and calls `comp.History(store, undo, redo)`.
14. **Wire includes**, For `rfw:"include:slotname"` on `*View` fields, calls `AddDependency(slotName, view)`.
15. **Resolve injects**, For `rfw:"inject"` fields, resolves from the DI container and sets the field.
16. **Validate**, Checks signal fields are pointer types and event handlers exist on the struct.

Returns the underlying `*core.HTMLComponent` (typed as `*View`).

---

## The Component Wrapper

`Component` wraps `*core.HTMLComponent` and exposes composition helpers:

```go
type Component struct {
    *core.HTMLComponent
}
```

### Prop

Registers a reactive signal under a key, making it available in the template:

```go
comp.Prop("count", composition.NewInt(0))
```

### On

Registers a handler function callable from the template via `@on:event:name`:

```go
comp.On("increment", func() { count.Set(count.Get() + 1) })
```

### Store

Creates or retrieves a store scoped to the component's ID:

```go
s := comp.Store("local", state.WithHistory(10))
s.Set("key", "value")
```

### History

Registers undo/redo handlers for a store:

```go
comp.History(s, "undo", "redo")
```

The handlers are registered in the DOM and can be used from templates with `@on:click:undo` / `@on:click:redo`.

### Unwrap

Access the underlying `*core.HTMLComponent` when needed:

```go
hc := comp.Unwrap()
```

---

## NewFrom\[T]()

Generic factory for zero-value struct types:

```go
view := composition.NewFrom[Counter]()
```

Equivalent to:

```go
view := composition.New(&Counter{})
```

Use it when you don't need custom field values during construction.

---

## FromProp\[T]

`FromProp` retrieves a signal from a component's props. If the prop is a plain value of type `T`, it wraps it into a new signal. If the prop is already a `Signal[T]`, it returns that signal directly.

```go
comp := composition.Wrap(core.NewHTMLComponent("card", tpl, map[string]any{"start": 5}))
sig := composition.FromProp[int](comp, "start", 0)
sig.Set(sig.Get() + 1) // start is now 6
```

If the prop key doesn't exist, a new signal with the default value is created and stored.

---

## Element Groups

Group multiple DOM elements for bulk operations:

```go
cards := composition.Group(
    composition.Div().Text("Card A"),
    composition.Div().Text("Card B"),
)
cards.AddClass("card").SetAttr("data-role", "item")
```

### Group Operations

| Method | Effect |
|--------|--------|
| `AddClass(name)` | Add a CSS class to all elements |
| `RemoveClass(name)` | Remove a CSS class |
| `ToggleClass(name)` | Toggle a CSS class |
| `SetAttr(name, value)` | Set an attribute |
| `SetStyle(prop, value)` | Set an inline style |
| `SetText(text)` | Set text content |
| `SetHTML(html)` | Replace inner HTML |
| `ForEach(fn)` | Iterate over elements |
| `Group(gs...)` | Merge with other groups |

### Node Builders

Create DOM elements programmatically:

```go
div := composition.Div().Class("box").Text("Hello")
span := composition.Span().Text("world")
btn := composition.Button().Text("Click me")
heading := composition.H(2).Text("Title")
link := composition.A().Href("/home").Text("Home")
```

All builders support `Class`, `Classes`, `Style`, `Styles`, `Text`, and `Group` methods. `A` adds `Href` and `Attr`.

---

## Bind and For

### Bind

Select an element by CSS selector and manipulate it:

```go
composition.Bind("#output", func(el composition.El) {
    el.Clear()
    el.Append(composition.Div().Text("Updated"))
})
```

### BindEl

Same as `Bind` but with a pre-selected element:

```go
composition.BindEl(domElement, func(el composition.El) {
    el.Append(composition.Span().Text("child"))
})
```

### For

Repeatedly call a function to generate nodes until it returns `nil`:

```go
composition.For("#list", func() composition.Node {
    if done {
        return nil
    }
    return composition.Div().Text("item")
})
```

---

## Includes

When a struct field is tagged `rfw:"include:slotname"` and holds a `*types.View`, `composition.New` automatically calls `AddDependency(slotname, view)` on the component. This wires the child view into the template at the `@include:slotname` position.

```go
type Page struct {
    composition.Component
    Header  *types.View `rfw:"include:header"`
    Content *types.View `rfw:"include:content"`
}
```

Set the fields before or after creating the view:

```go
page := &Page{}
page.Header = composition.New(&Header{})
page.Content = composition.New(&Content{})
view := composition.New(page)
```

If the field is `nil` at `composition.New` time, the include is skipped (no panic). Set it later and call `AddDependency` manually if needed.

---

## NewRaw

For layout or wrapper components that don't use `rfw` tags:

```go
view := composition.NewRaw("wrapper", tplBytes, map[string]any{"title": "Hello"})
```

`NewRaw` skips tag scanning entirely, it only initializes the HTMLComponent and default store.

---

## Type Aliases

The composition package re-exports signal and core types for convenience:

```go
type Int    = types.Int       // *state.Signal[int]
type String = types.String    // *state.Signal[string]
type Bool   = types.Bool       // *state.Signal[bool]
type Float  = types.Float     // *state.Signal[float64]
type Store  = types.Store     // *state.Store
type View   = types.View      // *core.HTMLComponent

var NewInt   = types.NewInt
var NewString = types.NewString
var NewBool  = types.NewBool
var NewFloat = types.NewFloat
```

---

## See Also

- [Components Basics](./components-basics), struct tags and component definition
- [Template Syntax](./template-syntax), RTML directives
- [Signals & Effects](./signals-effects-and-watchers), reactive state primitives