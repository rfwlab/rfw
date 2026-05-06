# markdown

```go
import "github.com/rfwlab/rfw/v2/markdown"
```

Markdown to HTML rendering using blackfriday.

## Parse

```go
func Parse(src string) string
```

Converts Markdown source to HTML.

## ExtractHeadings

```go
func ExtractHeadings(src string) []Heading
```

Extracts all headings from Markdown for table of contents.

| Field | Description |
| --- | --- |
| `Text` | Heading text |
| `Depth` | Heading level (1-6) |
| `ID` | Auto-generated slug ID |

## Example

```go
html := markdown.Parse("# Hello\n\nThis is **bold**.")
headings := markdown.ExtractHeadings("# Hello\n## World")
```