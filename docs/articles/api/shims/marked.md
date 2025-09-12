# marked

## Why
Provide simple helpers to work with the `marked` Markdown parser from Go without manual JavaScript calls.

## When to use
Use when a wasm application needs to parse or tokenize Markdown using the global `marked` library. Skip if markdown parsing happens elsewhere or the library is unavailable.

## How
1. Import the shim.
2. Call `Parse` or `Lexer` depending on your needs.

```go
import (
    marked "github.com/rfwlab/rfw/v1/js/shim/marked"
)

html := marked.Parse("# Title")
tokens := marked.Lexer("# Title")
```

## API
- `marked.Parse(src string) string`
- `marked.Lexer(src string) js.Value`

## Example
```go
import (
    marked "github.com/rfwlab/rfw/v1/js/shim/marked"
)

func render(md string) string {
    return marked.Parse(md)
}
```

## Notes
- The global `marked` object must be loaded before calling these helpers.
- Only available when targeting `js/wasm`.

## Related links
- [js](../js)
