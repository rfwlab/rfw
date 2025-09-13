# shortcut

Keyboard shortcut plugin binding key combinations to functions.

## Why

Map keyboard shortcuts to application functions without manual event wiring.

## Prerequisites

Register the plugin with the application:

```go
import shortcut "github.com/rfwlab/rfw/v1/plugins/shortcut"

core.RegisterPlugin(shortcut.New())
```

## How

1. Import the package.
2. Register handlers with `shortcut.Bind(combo, fn)`.
3. Press the matching keys to invoke the handler.

## API

| Function | Description |
| --- | --- |
| `shortcut.New()` | Construct the plugin. |
| `shortcut.Bind(combo, fn)` | Run `fn` when `combo` is pressed. |

## Example end-to-end

@include:ExampleFrame:{code:"/examples/components/shortcut_component.go", uri:"/examples/shortcut"}

## Notes and Limitations

- Key names are case-insensitive and sorted internally.
- Shortcuts fire repeatedly while keys remain pressed.

## Related links

- [events](events)
- [plugins](plugins)
