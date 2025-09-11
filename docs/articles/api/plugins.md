# plugins

| Method | Description |
| --- | --- |
| `PreBuild(cfg json.RawMessage) error` | Invoked before compilation begins. |
| `Build(cfg json.RawMessage) error` | Runs during the build step. |
| `PostBuild(cfg json.RawMessage) error` | Runs after the build completes. |
| `Install(a *core.App)` | Registers runtime features before the app starts. |
| `Uninstall(a *core.App)` | Cleans up resources added during `Install`. |
| `Name() string` | Unique identifier reported to the CLI. |
| `ShouldRebuild(path string) bool` | Signals if a file change triggers a rebuild. |
| `Priority() int` | Execution order – lower numbers run first. |

## Overview

The plugin system lets libraries extend rfw's compiler and runtime. Rather than
conforming to a strict interface, plugins only need to implement the hooks they
care about. The CLI automatically discovers methods by name and invokes them
when present.

## Build hooks

For build-time use a plugin must still provide a name, rebuild trigger and
priority. CLI build plugins are executed by priority with lower numbers running
first. Implement only the hooks your plugin needs.

```go
func (p *MyPlugin) Name() string
func (p *MyPlugin) ShouldRebuild(path string) bool
func (p *MyPlugin) Priority() int
```

## Runtime hooks

Plugins are registered with `core.RegisterPlugin(...)` before compiling the
WASM bundle. During `Install` they can register components, add routes or inject
scripts, while `Uninstall` can clean up any hooks.

## Store hooks

Plugins can react to store mutations by registering a store hook. The hook is
invoked for every `state.Set` call with the module, store, key and value:

```go
func (p loggerPlugin) Install(a *core.App) {
    a.RegisterStore(func(module, store, key string, value any) {
        // handle mutation
    })
}
```

Behind the scenes this uses `state.StoreHook`, allowing multiple plugins to
observe mutations without interfering with each other.

## Usage

Plugins must be registered before the application starts using
`core.RegisterPlugin`. During `Install` they can modify the app or add
features, and `Uninstall` can remove them.

Plugins can hook into the app lifecycle as illustrated here.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
