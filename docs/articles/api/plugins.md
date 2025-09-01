# plugins

The plugin system lets libraries extend rfw's compiler and runtime. Rather
than conforming to a strict interface, plugins only need to implement the
hooks they care about. The CLI automatically discovers methods by name and
invokes them when present:

```go
func (p *MyPlugin) PreBuild(cfg json.RawMessage) error { /* optional */ }
func (p *MyPlugin) Build(cfg json.RawMessage) error    { /* optional */ }
func (p *MyPlugin) PostBuild(cfg json.RawMessage) error { /* optional */ }
func (p *MyPlugin) Install(a *core.App)                { /* optional */ }
func (p *MyPlugin) Uninstall(a *core.App)              { /* optional */ }
```

For build-time use a plugin must still provide a name, rebuild trigger and
priority:

```go
func (p *MyPlugin) Name() string
func (p *MyPlugin) ShouldRebuild(path string) bool
func (p *MyPlugin) Priority() int
```

CLI build plugins are executed by priority with lower numbers running first.

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
