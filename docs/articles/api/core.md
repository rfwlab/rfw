# core

```go
import "github.com/rfwlab/rfw/v2/core"
```

Internal-facing runtime building blocks. Minimal public surface.

## Component Interface

| Item | Description |
| --- | --- |
| `Component` | Interface: `Init`, `Mount`, `Unmount`, `Render`. |

## HTMLComponent

| Function / Method | Description |
| --- | --- |
| `NewHTMLComponent(name, template string, props any) *HTMLComponent` | Create a view component. |
| `Init()` | Initialize the component. |
| `Mount()` | Lifecycle: called on first render. |
| `Unmount()` | Lifecycle: called on removal. |
| `Render() string` | Produce HTML output. |
| `AddDependency(dep Component)` | Register a child dependency. |
| `SetOnMount(fn func())` | Set the mount callback. |
| `SetOnUnmount(fn func())` | Set the unmount callback. |

## Utilities

| Function | Description |
| --- | --- |
| `SetDevMode(enabled bool)` | Toggle dev mode logging. |