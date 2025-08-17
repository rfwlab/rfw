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
