# Highlight Plugin

The **Highlight plugin** adds syntax highlighting for RTML and Go code blocks. It is included with rfw as a built-in plugin, so you don't need any external library.

## Features

* Highlights **RTML** (tags, attributes, variables, commands).
* Highlights **Go** code.
* Injects CSS automatically – no separate stylesheet required.
* Works both in Go and JavaScript contexts.

## Usage

1. Import the plugin.
2. Register it with your app.
3. Call `HighlightAll()` after rendering.

```go
import (
    highlight "github.com/rfwlab/rfw/v1/plugins/highlight"
    "github.com/rfwlab/rfw/v1/core"
)

func main() {
    core.RegisterPlugin(highlight.New())
    highlight.HighlightAll() // highlight all code blocks
}
```

In JavaScript, you can also use:

```js
rfwHighlight(code, lang) // returns highlighted HTML
rfwHighlightAll()        // highlights all code blocks
```

## API Reference

| Function                   | Description                                              |
| -------------------------- | -------------------------------------------------------- |
| `highlight.New()`          | Creates the plugin instance.                             |
| `highlight.HighlightAll()` | Highlights all `<pre><code>` blocks on the page.         |
| `rfwHighlight(code, lang)` | JS helper: highlight a string as `lang` and return HTML. |
| `rfwHighlightAll()`        | JS helper: run highlighting on all code blocks.          |

## Example Template

```html
<pre><code class="language-go">
func main() {
    println("Hello rfw!")
}
</code></pre>
```

With the plugin registered, the code block will be automatically highlighted.

## Notes

* Supported languages: **rtml**, **go**.
* Class names are case-insensitive (`language-go`, `language-Go`, etc.).
* Base styles are injected at runtime; override `.hl-*` classes to customize.
* Falls back to Highlight.js if `rfwHighlight` returns an empty string (deprecated soon).
