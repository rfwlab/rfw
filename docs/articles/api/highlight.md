# highlight

Custom syntax highlighting plugin implemented using rfw APIs.

## Why
Use this plugin to highlight RTML and Go code blocks without relying on external libraries. RTML support covers HTML tags, attributes, variables and commands.

## Prerequisites
Register the plugin with the application.

## How
1. Import the package.
2. Register it via `core.RegisterPlugin(highlight.New())`.
3. The plugin injects its CSS when installed; no separate stylesheet is needed.
4. Use the global `rfwHighlight` or `rfwHighlightAll` helpers to process code blocks.

## API

| Function | Description |
| --- | --- |
| `highlight.New()` | Construct the plugin. |
| `rfwHighlight(code, lang)` | Highlight `code` as `lang` and return HTML. |
| `rfwHighlightAll()` | Highlight all `<pre><code>` blocks on the page. |

## Example

```go
import (
    highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
    "github.com/rfwlab/rfw/v1/core"
)

func main() {
    core.RegisterPlugin(highlight.New())
}
```

```html
<script>
marked.setOptions({
    highlight: (code, lang) => rfwHighlight(code, lang),
});
</script>
```

## Notes and Limitations
- Supports only `rtml` and `go` languages.
- Falls back to Highlight.js when `rfwHighlight` returns an empty string.
- Injects base styles at runtime; override the `.hl-*` classes to customize.

## Related Links
- [plugins](../plugins)
- [js](../js)
- [dom](../dom)
- [highlightjs shim](shims/highlightjs)
