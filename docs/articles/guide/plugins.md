# Plugins

Plugins extend the **rfw** toolchain. They can add work before, during, or after a build, and they can also extend runtime behavior inside the app.

---

## Lifecycle Hooks

A plugin may implement any combination of lifecycle methods. The CLI discovers hooks by name and invokes them automatically:

```go
func (p *Plugin) PreBuild(cfg json.RawMessage) error  { /* optional */ }
func (p *Plugin) Build(cfg json.RawMessage) error     { /* optional */ }
func (p *Plugin) PostBuild(cfg json.RawMessage) error { /* optional */ }
func (p *Plugin) Install(a *core.App)                 { /* optional */ }
func (p *Plugin) Uninstall(a *core.App)               { /* optional */ }
```

Plugins that participate in builds also implement `Name`, `ShouldRebuild`, and `Priority` so the CLI can register and order them.

### PreBuild

Runs before compilation. Use it to prepare inputs or fetch resources. Example: download a schema or generate code consumed later in the build.

### Build

Runs while producing the Wasm bundle. Often compiles assets, copies files, or emits artifacts.

### PostBuild

Fires after the bundle is ready. Useful for minification, hash verification, or cleaning up temporary files.

### Install

Called at app startup. Use it to register routes, add components, or inject scripts—this extends runtime behavior.

### Uninstall

Runs when the app shuts down or the plugin is removed. Use it to detach watchers or undo `Install` changes.

---

## Priority and Ordering

Plugins can control execution order with `Priority() int`. Lower numbers run first.

```go
type first struct{}
func (p *first) Priority() int { return 0 }

type second struct{}
func (p *second) Priority() int { return 10 }

plugins.Register(&second{})
plugins.Register(&first{})
// "first" runs before "second" despite registration order.
```

---

## Writing a Plugin

A minimal plugin defines a type, implements hooks, and registers itself:

```go
package analytics

import (
    "encoding/json"
    "github.com/rfwlab/rfw/v1/core"
)

type Plugin struct{}

func New() *Plugin { return &Plugin{} }
func (p *Plugin) Name() string           { return "analytics" }
func (p *Plugin) Priority() int          { return 0 }
func (p *Plugin) ShouldRebuild(path string) bool { return false }

func (p *Plugin) PreBuild(cfg json.RawMessage) error { return nil }
func (p *Plugin) Build(cfg json.RawMessage) error    { return nil }
func (p *Plugin) PostBuild(cfg json.RawMessage) error { return nil }

func (p *Plugin) Install(a *core.App)   {}
func (p *Plugin) Uninstall(a *core.App) {}
```

Register in `main.go`:

```go
func main() {
    core.RegisterPlugin(analytics.New())
}
```

---

## Official Plugins

rfw ships with several plugins. Example: the `i18n` plugin for translations:

```go
import "github.com/rfwlab/rfw/v1/plugins/i18n"

core.RegisterPlugin(i18n.New(map[string]map[string]string{
    "en": {"hello": "Hello"},
    "it": {"hello": "Ciao"},
}))
```

---

## Use Cases

Plugins are ideal for:

* Analytics integration
* Asset processing
* Internationalization
* Feature flags

By choosing hooks carefully, a plugin can focus on build-time, runtime, or both.

@include\:ExampleFrame:{code:"/examples/plugins/plugins\_component.go", uri:"/examples/plugins"}
