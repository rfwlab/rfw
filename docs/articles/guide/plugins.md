# Plugins

Plugins let you hook into the rfw build and runtime pipeline. A plugin
implements the `Plugin` interface and is registered before compilation.

```go
type Plugin interface {
    Name() string
    Setup(*core.App) error
}
```

To create one, define a type and register it:

```go
package analytics

import "github.com/rfwlab/rfw/v1/core"

type Plugin struct{}

func (Plugin) Name() string { return "analytics" }

func (Plugin) Setup(app *core.App) error {
    // attach global scripts or modify the app
    return nil
}
```

In `main.go`:

```go
app := core.NewApp()
app.Use(analytics.Plugin{})
```

Plugins are ideal for analytics, custom elements or build-time
transformations.
Plugins extend the framework at runtime.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
