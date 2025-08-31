# core

The `core` package contains the application runtime and component
interfaces that everything else builds upon.

| Item | Description |
| --- | --- |
| `HTMLComponent` | Base struct for reactive components backed by RTML templates. |
| `Logger` | Interface for redirecting log output. |
| `ComponentRegistry` | Global registry for components loaded via `rt-is`. |

## Logger

Implement `core.Logger` and register with `core.SetLogger` to redirect log
output.

## Development mode

`core.SetDevMode(enabled bool)` toggles additional runtime checks and
warnings during development.

## Component helpers

`core.NewComponent(name, tpl, props)` returns an initialized `*core.HTMLComponent` with the provided template and props.
For structs embedding `*core.HTMLComponent`, use `core.NewComponentWith(name, tpl, props, self)` to bind lifecycle hooks without manual setup.

## Dynamic components

`core.ComponentRegistry` holds constructors for components that can be
loaded on demand. Register components with `core.RegisterComponent` and
retrieve them with `core.LoadComponent`:

```go
core.RegisterComponent("red-cube", func() core.Component {
        return NewRedCubeComponent()
})
comp := core.LoadComponent("red-cube")
```

Once registered, components can be rendered dynamically using the
`rt-is` attribute. It accepts either a static name or an expression:

```rtml
<div rt-is="red-cube"></div>
<div rt-is="{current}"></div>
```

The example below demonstrates dynamic component loading.

@include:ExampleFrame:{code:"/examples/components/dynamic_component.go", uri:"/examples/dynamic"}

## Dependency Injection

Components can share values without threading them through props. A parent
calls `Provide` to expose a value, while descendants call `Inject` or
`InjectTyped` to retrieve it. The `Inject` method returns an untyped `any`,
and the generic `core.Inject[T]` helper (alias `InjectTyped`) casts the value
to the requested type.

```go
parent := core.NewComponent("Parent", parentTpl, nil)
child := core.NewComponent("Child", childTpl, nil)

parent.Provide("answer", 42)
parent.AddDependency("child", child)

// Untyped lookup
v, _ := child.Inject("answer")

// Typed helper
answer, _ := core.Inject[int](child, "answer")
```

## Usage

Components are created with `core.NewComponent` by passing a name, template
and initial properties. Dependencies are added with `AddDependency`.

The example below shows how core components are composed.

@include:ExampleFrame:{code:"/examples/components/parent_component.go", uri:"/examples/parent"}
