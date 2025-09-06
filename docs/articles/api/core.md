# core

The `core` package contains the application runtime and component
interfaces that everything else builds upon.

| Item | Description |
| --- | --- |
| `HTMLComponent` | Base struct for reactive components backed by RTML templates. |
| `Logger` | Interface for redirecting log output. |
| `ComponentRegistry` | Global registry for components loaded via `rt-is`. |
| `ErrorBoundary` | Wrapper that renders fallback UI when a child panics. |

## Logger

Implement `core.Logger` and register with `core.SetLogger` to redirect log
output.

## Development mode

`core.SetDevMode(enabled bool)` toggles additional runtime checks and
warnings during development.

## Component helpers

`core.NewComponent(name, tpl, props)` returns an initialized `*core.HTMLComponent` with the provided template and props.
For structs embedding `*core.HTMLComponent`, use `core.NewComponentWith(name, tpl, props, self)` to bind lifecycle hooks without manual setup.

## Lifecycle hooks

Components expose entry points for code that should run when they are inserted into or removed from the DOM.

- `OnMount()` runs after the component's template is attached to the page.
- `OnUnmount()` executes just before the component is detached.
- `SetOnMount(fn)` and `SetOnUnmount(fn)` register hook functions on an `HTMLComponent` without defining methods on the struct.
- `WithLifecycle(child, mount, unmount)` wraps any component and attaches the provided mount and unmount functions.

```go
type Timer struct {
    *core.HTMLComponent
    stop func()
}

func NewTimer() *Timer {
    tpl := []byte("<span></span>")
    t := &Timer{HTMLComponent: core.NewComponent("Timer", tpl, nil)}
    t.SetOnMount(func(_ *core.HTMLComponent) {
        ticker := time.NewTicker(time.Second)
        t.stop = ticker.Stop
    })
    t.SetOnUnmount(func(_ *core.HTMLComponent) {
        if t.stop != nil {
            t.stop()
        }
    })
    return t
}
```

## ErrorBoundary

`core.NewErrorBoundary(child, fallback)` wraps a component and replaces its output with the provided fallback HTML when the child panics during `Render` or `Mount`.

```go
child := core.NewComponent("unsafe", tpl, nil)
safe := core.NewErrorBoundary(child, "<div>sorry</div>")
html := safe.Render() // renders fallback if child panics
```

## Dynamic components

`core.ComponentRegistry` holds constructors for components that can be
loaded on demand. Register components with `core.MustRegisterComponent`
(or `core.RegisterComponent` if you prefer explicit error handling) and
retrieve them with `core.LoadComponent`. `MustRegisterComponent` panics if a
component with the same name is already registered:

```go
core.MustRegisterComponent("red-cube", func() core.Component {
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

## Suspense

`Suspense` displays a fallback string while waiting for asynchronous
content. Create it with `core.NewSuspense(render func() (string,
error), fallback string) *core.Suspense`. `Render` executes the
function and keeps returning the fallback until the function stops
returning `http.ErrPending`; other errors are stringified.

```go
import (
        "fmt"

        "github.com/rfwlab/rfw/v1/core"
        "github.com/rfwlab/rfw/v1/http"
)

var todo Todo
content := core.NewSuspense(func() (string, error) {
        if err := http.FetchJSON("/api/todo/1", &todo); err != nil {
                return "", err
        }
        return fmt.Sprintf("<div>%s</div>", todo.Title), nil
}, "<div>Loading...</div>")
```

## Usage

Components are created with `core.NewComponent` by passing a name, template
and initial properties. Dependencies are added with `AddDependency`.

The example below shows how core components are composed.

@include:ExampleFrame:{code:"/examples/components/parent_component.go", uri:"/examples/parent"}
