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
4. Call `highlight.HighlightAll()` after rendering to process all code blocks.
   For JavaScript integrations, the global `rfwHighlight` and `rfwHighlightAll` helpers remain available. `HighlightAll` detects
   the language from `language-<lang>` classes (case-insensitive) or a `data-lang` attribute on `<code>` elements.

## API

| Function | Description |
| --- | --- |
| `highlight.New()` | Construct the plugin. |
| `highlight.HighlightAll()` | Highlight all `<pre><code>` blocks on the page. |
| `rfwHighlight(code, lang)` | Highlight `code` as `lang` and return HTML. |
| `rfwHighlightAll()` | Global JS wrapper calling `highlight.HighlightAll()`. |

## Example

```go
import (
    highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
    "github.com/rfwlab/rfw/v1/core"
)

func main() {
    core.RegisterPlugin(highlight.New())
    // highlight code blocks after the DOM is ready
    highlight.HighlightAll()
}
```

## Notes and Limitations
- Supports only `rtml` and `go` languages.
- Falls back to Highlight.js when `rfwHighlight` returns an empty string.
- Injects base styles at runtime; override the `.hl-*` classes to customize.
- `highlight.HighlightAll` (and the JS wrapper `rfwHighlightAll`) matches class names case-insensitively and checks the `data-lang` attribute when no class is present.

## Related Links
- [plugins](../plugins)
- [js](../js)
- [dom](../dom)
- [highlightjs shim](shims/highlightjs)
