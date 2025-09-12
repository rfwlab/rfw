# highlightjs

## Why
Registering Highlight.js languages from Go keeps syntax highlighting definitions within your Go sources and avoids extra script tags.

## When to use
Use this package when a wasm application needs custom Highlight.js languages. Skip it if built‑in languages from the CDN cover your needs or the environment lacks the global `hljs` object.

## How
1. Import the package.
2. Call `RegisterLanguage` with the language name and a definition callback.

```go
import (
    hljs "github.com/rfwlab/rfw/v1/js/shim/highlightjs"
    js "github.com/rfwlab/rfw/v1/js"
)

hljs.RegisterLanguage("rtml", func(h js.Value) js.Value {
    // construct the language definition
    def := js.NewDict()
    def.Set("name", "rtml")
    return def.Value
})
```

## API
- `highlightjs.RegisterLanguage(name string, def func(hljs js.Value) js.Value)`

## Example
The docs site registers an `rtml` language at start‑up:

```go
func init() {
    hljs.RegisterLanguage("rtml", func(h js.Value) js.Value {
        xml := h.Call("getLanguage", "xml")
        reg := js.Get("RegExp")
        interpolation := js.NewDict()
        interpolation.Set("className", "template-variable")
        interpolation.Set("begin", reg.New("\\{"))
        interpolation.Set("end", reg.New("\\}"))
        interpolation.Set("relevance", 0)
        arr := js.NewArray()
        arr.Push(interpolation.Value)
        contains := js.ArrayOf(xml.Get("contains")).Concat(arr).Value
        def := js.NewDict()
        def.Set("contains", contains)
        return h.Call("inherit", xml, def.Value)
    })
}
```

## Notes
- The global `hljs` object must be loaded before registration.
- Language definitions require manual construction of JavaScript values.
- Only available when building for `js/wasm`.

## Related links
- [js](../js)
- [bundler plugin](../bundler-plugin)
