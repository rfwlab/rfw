# markdown

Convert Markdown content to HTML directly from Go without relying on external JavaScript libraries.

| Function | Description |
| --- | --- |
| `Parse(src string) string` | Convert `src` Markdown to HTML. |
| `Headings(src string) []markdown.Heading` | Return headings with text, depth and id. |

