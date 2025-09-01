# plugins

The plugin system lets libraries extend rfw's compiler and runtime. A
plugin implements the core interface:

```go
type Plugin interface {
    Build(json.RawMessage) error
    Install(*core.App)
}
```

Plugins may optionally hook into additional lifecycle stages:

```go
type PreBuilder interface { PreBuild(json.RawMessage) error }
type PostBuilder interface { PostBuild(json.RawMessage) error }
type Uninstaller interface { Uninstall(*core.App) }
```

CLI build plugins are executed by priority (lower numbers run first).

Plugins are registered with `core.RegisterPlugin(...)` before compiling the WASM
bundle. During `Install` they can register components, add routes or inject
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
