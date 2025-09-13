# seo

## Context
Manages the document `<title>` and `<meta>` tags to improve search engine optimisation.

## Prerequisites
Use when an application needs to update `<title>` or `<meta>` elements dynamically.

## How
1. Enable the plugin in `rfw.json`:
```json
{
  "plugins": {
    "seo": {
      "title": "RFW",
      "pattern": "%s | RFW",
      "meta": {
        "description": "RFW documentation"
      }
    }
  }
}
```
2. Register the plugin before startup:
```go
import "github.com/rfwlab/rfw/v1/core"
import "github.com/rfwlab/rfw/v1/plugins/seo"

func init() { core.RegisterPlugin(seo.New()) }
```
3. Update the values at runtime:
```go
seo.SetTitle("About")
seo.SetMeta("description", "Page description")
```

## API
- `func SetTitle(title string)`
- `func SetMeta(name, content string)`

## Example
```go
p := seo.New()
_ = p.Build([]byte(`{"title":"RFW","pattern":"%s | RFW"}`))
p.Install(nil)
seo.SetTitle("Docs")
seo.SetMeta("description", "RFW documentation")
```

## Notes and Limitations
Creates missing `<title>` and `<meta>` elements in the document `<head>`. The `pattern` uses `fmt.Sprintf` semantics and should include a single `%s` placeholder.

## Related links
- [plugins](plugins)
- [dom](../dom)
