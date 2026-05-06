# monitor

```go
import "github.com/rfwlab/rfw/v2/plugins/monitor"
```

DOM mutation and intersection observer plugin.

## Plugin

```go
type Plugin struct {
    MutationSelector     string
    IntersectionSelector string
    IntersectionOpts     js.Value
    Mutations     chan js.Value
    Intersections chan js.Value
}
```

| Field | Description |
| --- | --- |
| `MutationSelector` | CSS selector to observe |
| `IntersectionSelector` | CSS selector for intersection observer |
| `IntersectionOpts` | IntersectionObserver options |
| `Mutations` | Channel of mutation records |
| `Intersections` | Channel of intersection entries |

## Constructor

```go
func New(mSel, iSel string, opts js.Value) *Plugin
```

- `mSel`: selector for MutationObserver
- `iSel`: selector for IntersectionObserver  
- `opts`: IntersectionObserver options (threshold, root, rootMargin)

## Usage

```go
plugin := monitor.New(".dynamic-content", ".visible", js.Global().Get("IntersectionObserver"))
a.RegisterPlugin(plugin)

for mutation := range plugin.Mutations {
    fmt.Println("Modified:", mutation)
}
```