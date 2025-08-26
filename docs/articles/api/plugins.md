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

Plugins are registered with `app.Use(...)` before compiling the WASM
bundle. During `Install` they can register components, add routes or inject
scripts, while `Uninstall` can clean up any hooks.

## Usage

Plugins must be registered before the application starts using
`core.RegisterPlugin` or `app.Use`. During `Install` they can modify the app or
add features, and `Uninstall` can remove them.

Plugins can hook into the app lifecycle as illustrated here.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
