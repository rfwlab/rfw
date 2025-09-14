# Component Composition

## Context / Why

The `composition` package offers a thin wrapper around an existing `*core.HTMLComponent`.
Use it when you need to embed that component inside a typed struct without changing its behavior.

## Prerequisites / When

Use this package when you already have a `*core.HTMLComponent` created elsewhere and want to expose it as a dedicated type.

## How

1. Create or obtain a `*core.HTMLComponent`.
2. Pass it to `composition.Wrap` to get a `*composition.Component`.
3. Call `Unwrap` to access the original component when needed.

```go
import (
    "github.com/rfwlab/rfw/v1/composition"
    core "github.com/rfwlab/rfw/v1/core"
)

hc := core.NewComponent("Name", nil, nil)
cmp := composition.Wrap(hc)
```

## Props

Expose reactive state as component properties. `Prop` stores a `state.Signal`
under a key, while `FromProp` retrieves an existing signal or creates a new one
with a default value. If a prop holds a plain value matching the requested
type, `FromProp` wraps it in a new signal; incompatible types panic. `Prop`
never overwrites an existing plain prop—use `FromProp` to obtain a synchronized
signal when a non-reactive value is present.

```go
cmp := composition.Wrap(core.NewComponent("Counter", nil, nil))
count := state.NewSignal(0)
cmp.Prop("count", count)

other := composition.FromProp[int](cmp, "other", 1)
other.Set(2)
```

```go
// Legacy prop plain -> FromProp wraps without altering Props
hc := core.NewComponent("Counter", nil, map[string]any{"count": 5})
cmp := composition.Wrap(hc)
count := composition.FromProp[int](cmp, "count", 0) // -> 5, Props["count"] remains 5 (int)
count.Set(6)
```

## Event Handlers

Attach functions to DOM events by name. The wrapper forwards handlers to the
`dom` package:

```go
cmp := composition.Wrap(core.NewComponent("Counter", nil, nil))
cmp.On("save", func() { /* handle save */ })
```

## Example

```go
hc := core.NewComponent("Hello", nil, nil)
wrapped := composition.Wrap(hc)
_ = wrapped.Unwrap().Render()
```

## Notes and Limitations

`composition.Component` adds no new features; it only encapsulates the existing HTML component.

## Related Links

- [Component Basics](./components-basics)

## DOM Bindings

### Context / Why

Manual components sometimes need to manipulate existing DOM nodes. Annotating a
template element with a constructor such as `[list]` lets the component fetch it
directly with `GetRef`. `Bind` and `For` remain available for ad‑hoc selector
based access without touching `syscall/js`.

### Prerequisites / When

Use when a component must programmatically render plain nodes outside the
template system.

### How

1. Place a constructor like `[list]` on the target element in your template.
2. Call `GetRef("list")` on the component to obtain the `dom.Element`.
3. Clear or append children using `SetHTML` and `AppendChild` or builders like
   `Div().Class("c").Text("hi")`, `Span().Text("hi")`, or `Button().Text("go")`.
4. Alternatively, `Bind` and `For` accept CSS selectors and perform similar
   operations when a template ref isn't available.

```go
tpl := []byte("<root><div [list]></div></root>")
cmp := composition.Wrap(core.NewComponent("List", tpl, nil))
items := []string{"a", "b"}
cmp.SetOnMount(func(*core.HTMLComponent) {
    listEl := cmp.GetRef("list")
    listEl.SetHTML("")
    for _, item := range items {
        listEl.AppendChild(composition.Div().Text(item).Element())
    }
})
```

### Example

```go
tpl := []byte("<root><div [greet]></div></root>")
cmp := composition.Wrap(core.NewComponent("Greet", tpl, nil))
cmp.SetOnMount(func(*core.HTMLComponent) {
    el := cmp.GetRef("greet")
    el.SetHTML("")
    el.AppendChild(composition.Div().Text("hello").Element())
})
```

### Notes and Limitations

- `For` stops when the generator returns `nil`.
- Missing selectors are ignored silently.

### Related Links

- [dom](../api/dom)

## Element Groups

### Context / Why

Handling several nodes created via the composition builders often requires
manual loops. `Group` collects those nodes and provides helpers to mutate all
of them at once.

### Prerequisites / When

Use when multiple elements created with constructors need similar updates
outside the template system.

### How

1. Create an empty group with `composition.NewGroup` or collect nodes with `composition.Group`.
2. Build nodes with constructors like `Div()` and add them to groups via the `Group` method.
3. Merge groups with `(composition.Elements).Group` when needed.
4. Apply bulk operations such as `AddClass`, `SetAttr`, or iterate with `ForEach`.

```go
import (
    "github.com/rfwlab/rfw/v1/composition"
    "github.com/rfwlab/rfw/v1/dom"
)

cards := composition.Group(
    composition.Div().Class("card"),
    composition.Div().Class("card"),
)
cards.AddClass("active").SetStyle("color", "red")
```

### Example

```go
cards := composition.Group(
    composition.Div().Text("A"),
    composition.Div().Text("B"),
)
cards.AddClass("card").SetAttr("data-role", "item")
```

### Notes and Limitations

- `Group` panics when called without nodes.
- `ForEach` panics if the callback is nil.
- Selectors and IDs are ignored; pass nodes explicitly.

### Related Links

- [dom](../api/dom)

## Stores and History

### Context / Why

Some composed components need isolated state. `Store` creates a namespaced
`state.Store` tied to the component, while `History` exposes undo/redo handlers
for that store.

### Prerequisites / When

Use when component logic requires local state with optional mutation history.

### How

1. Call `Store` with a name to create a component-scoped store.
2. Enable history via `state.WithHistory` if undo/redo is needed.
3. Register undo and redo handlers with `History`, scoping handler names with the component ID to avoid collisions, and reference them in the template.

```go
cmp := composition.Wrap(core.NewComponent("Counter", nil, nil))
s := cmp.Store("count", state.WithHistory(10))
s.Set("v", 1)
cmp.History(s, cmp.ID+":undo", cmp.ID+":redo")
```

### APIs Used

- `(*composition.Component).Store(name string, opts ...state.StoreOption) *state.Store`
- `(*composition.Component).History(s *state.Store, undo, redo string)`
- `state.WithHistory(limit int) state.StoreOption`
- `dom.RegisterHandlerFunc(name string, fn func())`

### End-to-End Example

```go
cmp := composition.Wrap(core.NewComponent("Counter", nil, nil))
s := cmp.Store("count", state.WithHistory(5))
s.Set("v", 1)
s.Set("v", 2)
cmp.History(s, cmp.ID+":undo", cmp.ID+":redo")
dom.GetHandler(cmp.ID + ":undo").Invoke() // -> v = 1
dom.GetHandler(cmp.ID + ":redo").Invoke() // -> v = 2
```

### Notes and Limitations

Undo/redo handlers work only when the store was created with `state.WithHistory`. Handler names live in a global registry; prefix them with the component ID to prevent collisions.

### Related Links

- [State history](../api/state#history)
- [DOM handlers](../api/dom#usage)
