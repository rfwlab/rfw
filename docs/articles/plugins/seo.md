# SEO Plugin

The **SEO plugin** manages the document `<title>` and `<meta>` tags. Use it when you want your rfw app to update these values dynamically, without manually touching the DOM.

## Features

* Set the document title with a pattern (`%s | rfw`, etc.).
* Add or update `<meta>` tags (e.g. description).
* Automatically creates missing `<title>` or `<meta>` elements.

## Setup

Enable the plugin in `rfw.json`:

```json
{
  "plugins": {
    "seo": {
      "title": "rfw",
      "pattern": "%s | rfw",
      "meta": {
        "description": "rfw documentation"
      }
    }
  }
}
```

Register it in your app:

```go
import "github.com/rfwlab/rfw/v1/core"
import "github.com/rfwlab/rfw/v1/plugins/seo"

func init() {
    core.RegisterPlugin(seo.New())
}
```

## Usage

Update values at runtime:

```go
seo.SetTitle("About")
seo.SetMeta("description", "Page description")
```

## API Reference

| Function                            | Description                                                      |
| ----------------------------------- | ---------------------------------------------------------------- |
| `seo.SetTitle(title string)`        | Updates the `<title>` element. Applies the configured `pattern`. |
| `seo.SetMeta(name, content string)` | Updates or creates a `<meta name="...">` element.                |

## Example

```go
p := seo.New()
_ = p.Build([]byte(`{"title":"rfw","pattern":"%s | rfw"}`))
p.Install(nil)
seo.SetTitle("Docs")
seo.SetMeta("description", "rfw documentation")
```

## Notes

* The `pattern` must contain a `%s` placeholder; it is passed to `fmt.Sprintf`.
* `SetTitle` and `SetMeta` can be reassigned before `core.RegisterPlugin` to customize behavior.
* Missing `<title>` or `<meta>` tags will be created automatically.
