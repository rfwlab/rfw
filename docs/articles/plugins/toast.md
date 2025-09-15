# Toast Plugin

The **Toast plugin** shows temporary notifications stacked in a corner of the screen. Each toast has a dismiss button and can include optional action buttons.

## Features

* Quick feedback messages for background actions.
* Automatic removal after a timeout.
* Optional action buttons and custom templates.
* Stacked layout with dismiss button by default.

## Setup

Register the plugin before starting your app:

```go
import (
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/plugins/toast"
)

func main() {
    core.RegisterPlugin(toast.New())
}
```

## Usage

Push a notification:

```go
toast.Push("Saved!")
```

Custom duration:

```go
toast.PushTimed("Short message", 1500)
```

With options:

```go
toast.PushOptions("Upload complete", toast.Options{
    Duration: 5000,
    Actions: []toast.Action{{Label: "Undo", Handler: func(){ /* ... */ }}},
})
```

## API Reference

| Function                                      | Description                                    |
| --------------------------------------------- | ---------------------------------------------- |
| `toast.New()`                                 | Creates the plugin with a 3s default duration. |
| `toast.Push(msg string)`                      | Shows a message for the default duration.      |
| `toast.PushTimed(msg string, d int)`          | Shows a message for `d` ms.                    |
| `toast.PushOptions(msg string, opts Options)` | Shows a message with custom options.           |

### Types

* `toast.Action` – defines an extra button with a label and handler.
* `toast.Options` – configures `Duration`, `Actions`, `Template`.

## Notes

* Messages stack and disappear after their duration or when dismissed.
* Inline styles are used for positioning.
* Override the default template with `Options.Template` for full customization.
