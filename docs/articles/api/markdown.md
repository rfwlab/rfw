# markdown

## Why
Convert Markdown content to HTML directly from Go without relying on external JavaScript libraries.

## Prerequisites
Use when your application needs to render Markdown at runtime. The package runs anywhere Go does.

## How
1. Import the package.
2. Call `markdown.Parse` to convert Markdown to HTML.
3. Use `markdown.Headings` to extract heading information.

## API
| Function | Description |
| --- | --- |
| `Parse(src string) string` | Convert `src` Markdown to HTML. |
| `Headings(src string) []markdown.Heading` | Return headings with text, depth and id. |

## Example
```go
import (
    "fmt"
    markdown "github.com/rfwlab/rfw/v1/markdown"
)

func render(md string) {
    html := markdown.Parse(md)
    fmt.Println(html)
    hs := markdown.Headings(md)
    fmt.Println(hs)
}
```

## Notes and Limitations
- Raw HTML in the source is preserved.
- Heading ids follow a simple slug algorithm and may change if text repeats.

## Related links
- [docs plugin](docs-plugin)
