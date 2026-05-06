# Composition

The `composition` package is the core of rfw v2's component system. `composition.New` uses **type-based auto-wiring** — no tags required. Struct field types determine everything: signals, stores, refs, injects, histories, and host bindings.

---

## How `composition.New` Works

`composition.New(&MyStruct{})` performs these steps:

1. **Scan struct fields** — the `scan` package inspects field types (not tags) to detect signals, stores, refs, injects, histories, host fields, and includes.
2. **Resolve template** — checks for an optional `Template() string` method, then falls back to convention (`StructName.rtml` or `StructName.html`).
3. **Create HTML component** — calls `core.NewHTMLComponent` with the resolved template.
4. **Wrap as Component** — `Wrap(hc)` creates a `composition.Component` providing `Prop`, `On`, `Store`, and `History` methods.
5. **Initialize default store** — creates or reuses the `"default"` store under the `"app"` module.
6. **Wire signals** — for each signal-type field (`t.Int`, `t.String`, etc.): if nil, auto-initializes with a zero-value signal; registers as a prop.
7. **Wire host signals** — `t.HInt`, `t.HString`, etc. are both signals and host component declarations.
8. **Wire stores** — `*t.Store` fields get the store from the global manager and registered on the component.
9. **Wire histories** — `*t.History` fields are bound to the first component store for undo/redo.
10. **Wire includes** — `*t.View` fields call `AddDependency(slotName, view)` using the lowercase field name.
11. **Wire injects** — `*t.Inject[T]` fields are resolved from the DI container by lowercase field name.
12. **Wire refs** — `*t.Ref` fields are allocated and resolved from the DOM on mount via `GetRef`.
13. **Auto-discover methods** — exported no-arg methods become event handlers (excluding `OnMount`/`OnUnmount` and Component methods).
14. **Wire lifecycle** — `OnMount` and `OnUnmount` are registered; `OnMount` also resolves ref DOM nodes.

Returns `(*View, error)`. On failure (no template, invalid type), returns a descriptive error instead of panicking.

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
view, err := composition.NewFrom[Counter]()
if err != nil {
    log.Fatal(err)
}
```

Equivalent to:

```go
view, err := composition.New(&Counter{})
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

When a struct field is `*types.View`, `composition.New` automatically calls `AddDependency(slotname, view)` where the slot name is the lowercase field name:

```go
type Page struct {
    composition.Component
    Header  *types.View
    Content *types.View
}
```

Set the fields before or after creating the view:

```go
page := &Page{}
page.Header = headerView
page.Content = contentView
view, err := composition.New(page)
```

If the field is `nil` at `composition.New` time, the include is skipped (no panic). Set it later and call `AddDependency` manually if needed.

---

## NewRaw

For layout or wrapper components that don't need type-based wiring:

```go
view := composition.NewRaw("wrapper", tplBytes, map[string]any{"title": "Hello"})
```

`NewRaw` skips scanning entirely, it only initializes the HTMLComponent and default store.

---

## Type Aliases

The composition package re-exports signal and core types for convenience:

```go
type Int    = types.Int       // *state.Signal[int]
type String = types.String    // *state.Signal[string]
type Bool   = types.Bool      // *state.Signal[bool]
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

- [Components Basics](./components-basics), struct fields and type-based wiring
- [Template Syntax](./template-syntax), RTML directives
- [Signals & Effects](./signals-effects-and-watchers), reactive state primitives