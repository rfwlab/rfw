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

## APIs Used

- `core.NewComponent(name string, templateFS []byte, props map[string]any) *core.HTMLComponent`
- `composition.Wrap(c *core.HTMLComponent) *composition.Component`
- `(*composition.Component).Unwrap() *core.HTMLComponent`

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

