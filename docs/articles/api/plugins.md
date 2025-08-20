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

## Example

```go
core.RegisterPlugin(logger.New())
core.RegisterPlugin(i18n.New(map[string]map[string]string{
        "en": {"hello": "Hello"},
        "it": {"hello": "Ciao"},
}))
core.RegisterPlugin(mon.New())
```

1. Multiple plugins are registered before the app starts.
2. The `logger` plugin intercepts logs while `i18n` provides translations.
3. `mon.New()` installs an additional monitoring plugin.
