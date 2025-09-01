# Plugins

Plugins extend the rfw toolchain. They can contribute work before, during or after the CLI build and modify the runtime application.

## Lifecycle hooks

The core `Plugin` interface requires `Build` and `Install` methods. Additional optional hooks let a plugin participate in more stages:

```go
type Plugin interface {
    Build(json.RawMessage) error    // run during CLI build
    Install(*core.App)              // configure the app at runtime
}

type PreBuilder interface { PreBuild(json.RawMessage) error }
type PostBuilder interface { PostBuild(json.RawMessage) error }
type Uninstaller interface { Uninstall(*core.App) }
```

### PreBuild
`PreBuild` executes before the CLI starts compiling. Use it to prepare input files or read configuration. For example, a plugin could download an external schema or generate source code that the subsequent build step consumes.

### Build
`Build` runs while the CLI is producing the wasm bundle. Typical implementations compile assets, copy files or emit additional artifacts. It receives any plugin specific configuration as raw JSON.

### PostBuild
`PostBuild` fires after the bundle is ready. It is a good place to minify output, verify hashes or remove temporary files created earlier.

### Install
`Install` runs when the application bootstraps. The plugin receives a `*core.App` and can register routes, components or inject scripts. This is where runtime behaviour is extended.

### Uninstall
`Uninstall` is invoked when tearing down the app or removing the plugin. Use it to detach watchers, delete generated files or undo registrations made in `Install`.

## Priority and ordering
Plugins may implement `Priority() int` to control execution order. Lower numbers run first, so a plugin returning `0` will precede one returning `10`. This matters when one plugin depends on the output of another.

The CLI sorts registered plugins by priority:

```go
type first struct{}
func (p *first) Priority() int { return 0 }

type second struct{}
func (p *second) Priority() int { return 10 }

plugins.Register(&second{})
plugins.Register(&first{})
// During the build step "first" runs before "second" despite registration order.
```

## Writing a plugin
A minimal plugin defines a type, implements the desired hooks and registers itself:

```go
package analytics

import (
    "encoding/json"

    "github.com/rfwlab/rfw/v1/core"
)

type Plugin struct{}

func New() core.Plugin { return &Plugin{} }

func (p *Plugin) Priority() int { return 0 }

func (p *Plugin) PreBuild(cfg json.RawMessage) error {
    // e.g. generate tracking configuration
    return nil
}

func (p *Plugin) Build(cfg json.RawMessage) error {
    // compile assets or copy files
    return nil
}

func (p *Plugin) PostBuild(cfg json.RawMessage) error {
    // cleanup temporary files
    return nil
}

func (p *Plugin) Install(a *core.App) {
    // inject analytics script into pages
}

func (p *Plugin) Uninstall(a *core.App) {
    // remove analytics script
}
```

Register the plugin in `main.go`:

```go
func main() {
    core.RegisterPlugin(analytics.New())
}
```

## Official plugins

RFW ships with a few ready-made plugins. The `i18n` package adds
basic string translation helpers:

```go
import "github.com/rfwlab/rfw/v1/plugins/i18n"

core.RegisterPlugin(i18n.New(map[string]map[string]string{
    "en": {"hello": "Hello"},
    "it": {"hello": "Ciao"},
}))
```

## Use cases
Plugins excel at tasks such as analytics integration, asset processing, internationalization or feature flagging. By choosing which hooks to implement, a plugin can focus on build-time concerns, runtime behaviour or both.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
