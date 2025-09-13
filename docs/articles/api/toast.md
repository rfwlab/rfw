# toast

A simple plugin that displays temporary notifications stacked in the corner.

## Why

Use this plugin to inform users about actions that complete in the background,
without disrupting the current view.

## Prerequisites

Register the plugin before starting the application:

```go
import (
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/plugins/toast"
)

func main() {
    core.RegisterPlugin(toast.New())
}
```

## How

1. Import the package.
2. Register the plugin with `core.RegisterPlugin`.
3. Invoke `toast.Push("message")` from handlers to enqueue notifications.

## API

| Function | Description |
| --- | --- |
| `New()` | Construct the plugin with a 3s default duration. |
| `Push(msg)` | Enqueue `msg` with the default duration. |
| `PushTimed(msg, d)` | Enqueue `msg` for `d` instead of the default. |

## Example

@include:ExampleFrame:{code:"/examples/plugins/toast_component.go", uri:"/examples/toast"}

## Notes and Limitations

Messages are shown sequentially and removed after their duration. Inline styles
are used for positioning; customize the appearance by overriding classes.

## Related Links

- [dom](../dom)
- [plugins](../plugins)

