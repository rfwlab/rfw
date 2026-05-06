# shortcut

```go
import "github.com/rfwlab/rfw/v2/plugins/shortcut"
```

Keyboard shortcut plugin for rfw.

## Plugin

```go
type Plugin struct {
    bindings map[string]func()
    pressed  map[string]bool
}
```

| Method | Description |
| --- | --- |
| `New() *Plugin` | Create a new plugin |
| `Build(json.RawMessage) error` | Build from config (no-op) |
| `Install(*core.App)` | Register key listeners |

## Bind

```go
func Bind(combo string, fn func())
```

Bind a keyboard combo to a handler. Combo format: `"control+k"`, `"shift+escape"`, etc.

## Example

```go
// In plugin config or code:
shortcut.Bind("control+s", func() {
    fmt.Println("Save!")
})

shortcut.Bind("control+shift+p", func() {
    fmt.Println("Secret!")
})
```

Combinations are normalized (order-independent, case-insensitive).