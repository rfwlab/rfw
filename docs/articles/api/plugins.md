# plugins

The plugin system lets libraries extend rfw's compiler and runtime. A
plugin implements:

```go
type Plugin interface {
    Name() string
    Setup(*core.App) error
}
```

Plugins are registered with `app.Use(...)` before compiling the WASM
bundle. During `Setup` they can register components, add routes or inject
scripts.

## Usage

Plugins must be registered before the application starts using
`core.RegisterPlugin` or `app.Use`. During `Setup` they can modify the app or
add features.

Plugins can hook into the app lifecycle as illustrated here.

@include:ExampleFrame:{code:"/examples/plugins/plugins_component.go", uri:"/examples/plugins"}
