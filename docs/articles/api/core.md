# core

The `core` package contains the application runtime and component
interfaces that everything else builds upon.

| Type | Description |
| --- | --- |
| `App` | Root object that mounts components and orchestrates routing. |
| `HTMLComponent` | Base struct for reactive components backed by RTML templates. |
| `Logger` | Interface for redirecting log output. |

## Logger

Implement `core.Logger` and register with `core.SetLogger` to redirect log
output.

## App

`core.NewApp(rootID string)` creates an application. Components are
mounted with `app.Mount(component)`.

## Component helpers

`core.NewComponent(name, tpl, props)` returns an initialized `*core.HTMLComponent` with the provided template and props.
For structs embedding `*core.HTMLComponent`, use `core.NewComponentWith(name, tpl, props, self)` to bind lifecycle hooks without manual setup.

## Usage

Components are created with `core.NewComponent` by passing a name, template
and initial properties. Dependencies are added with `AddDependency`.

## Example

```go
c := core.NewComponent("MainComponent", mainComponentTpl, nil)
headerComponent := NewHeaderComponent(map[string]any{
        "title": "Main Component",
})
c.AddDependency("header", headerComponent)
```

1. `core.NewComponent` instantiates the component with its template.
2. `NewHeaderComponent` creates a dependency to inject into the main
   component.
3. `AddDependency` makes the dependency available in the template under the
   name `header`.
