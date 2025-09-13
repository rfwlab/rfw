# shims

## Why
Shims provide minimal wrappers around browser libraries so Go code can call them without dealing with raw JavaScript bindings.

## When to use
Use a shim when your wasm module needs to interact with a JavaScript library that lacks first-class Go support. Skip them if standard APIs suffice.

## How
1. Import the required shim package.
2. Invoke the helper exported by the shim.

## Example
```go
import (
    js "github.com/rfwlab/rfw/v1/js"
    hljs "github.com/rfwlab/rfw/v1/js/shim/highlightjs"
)

func init() {
    hljs.RegisterLanguage("rtml", func(h js.Value) js.Value {
        def := js.NewDict()
        def.Set("name", "rtml")
        return def.Value
    })
}
```

## Notes
- Each shim requires its corresponding JavaScript library to be loaded.
- Shims are available only when targeting `js/wasm`.

## Related links
 - [js](../js)
 - [highlightjs](highlightjs)
