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
cmp.Prop("count", count) // Since: Unreleased

other := cmp.FromProp("other", 1) // Since: Unreleased
other.Set(2)
```

```go
// Legacy prop plain -> FromProp wraps without altering Props
hc := core.NewComponent("Counter", nil, map[string]any{"count": 5})
cmp := composition.Wrap(hc)
count := cmp.FromProp[int]("count", 0) // -> 5, Props["count"] remains 5 (int)
count.Set(6)
```

Both helpers are available since *Unreleased*.

## Event Handlers

Attach functions to DOM events by name. The wrapper forwards handlers to the
`dom` package:

```go
cmp := composition.Wrap(core.NewComponent("Counter", nil, nil))
cmp.On("save", func() { /* handle save */ }) // Since: Unreleased
```

## APIs Used

- `core.NewComponent(name string, templateFS []byte, props map[string]any) *core.HTMLComponent`
- `composition.Wrap(c *core.HTMLComponent) *composition.Component`
- `(*composition.Component).Unwrap() *core.HTMLComponent`
- `(*composition.Component).On(name string, fn func())`
- `(*composition.Component).Prop[T any](key string, sig *state.Signal[T])`
- `(*composition.Component).FromProp[T any](key string, def T) *state.Signal[T]`
- `state.NewSignal(initial T) *state.Signal[T]`

## End-to-End Example

```go
hc := core.NewComponent("Hello", nil, nil)
wrapped := composition.Wrap(hc)
_ = wrapped.Unwrap().Render()
```

## Notes and Limitations

`composition.Component` adds no new features; it only encapsulates the existing HTML component.

## Related Links

- [Component Basics](./components-basics)

