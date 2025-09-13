# toast

A simple plugin that displays temporary notifications stacked in the corner.
Each toast includes a dismiss button and can expose optional action buttons.

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
3. Invoke `toast.Push("message")` for a simple notification or
   `toast.PushOptions("message", opts)` to add buttons or a custom template.

## API

| Function | Description |
| --- | --- |
| `New()` | Construct the plugin with a 3s default duration. |
| `Push(msg)` | Enqueue `msg` with the default duration. |
| `PushTimed(msg, d)` | Enqueue `msg` for `d` instead of the default. |
| `PushOptions(msg, opts)` | Enqueue `msg` with custom `Options`. |
| `Action{Label, Handler}` | Defines an additional button. |
| `Options{Duration, Actions, Template}` | Configures a toast. |

## Example

@include:ExampleFrame:{code:"/examples/plugins/toast_component.go", uri:"/examples/toast"}

## Notes and Limitations

Messages are shown sequentially and removed after their duration or when the
dismiss button is pressed. Inline styles are used for positioning. Override the
default template or supply a custom one through `Options.Template` to fully
customize the appearance.

## Related Links

- [dom](../dom)
- [plugins](../plugins)

