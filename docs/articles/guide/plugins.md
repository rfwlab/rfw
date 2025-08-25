# Plugins

Plugins let you hook into the rfw build and runtime pipeline. A plugin
implements the `Plugin` interface and is registered before compilation.

```go
type Plugin interface {
    Build(json.RawMessage) error
    Install(*core.App)
}
```

To create one, define a type and register it:

```go
package analytics

import (
    "encoding/json"

    "github.com/rfwlab/rfw/v1/core"
)

type Plugin struct{}

func New() core.Plugin { return &Plugin{} }

func (p *Plugin) Build(cfg json.RawMessage) error {
    // optional build-time work using cfg
    return nil
}

func (p *Plugin) Install(a *core.App) {
    // attach global scripts or modify the app
}
```

In `main.go`:

```go
func main() {
    core.RegisterPlugin(analytics.New())
}
```

Plugins are ideal for analytics, custom elements or build-time
transformations. During `Install` plugins can register hooks on
`*core.App` such as `RegisterRouter`, `RegisterStore`, `RegisterTemplate`
or `RegisterLifecycle` to extend the framework at runtime.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
