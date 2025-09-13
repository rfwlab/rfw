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
      "title": "rfw",
      "pattern": "%s | rfw",
      "meta": {
        "description": "rfw documentation"
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
- `var SetTitle func(title string)`
- `var SetMeta func(name, content string)`

## Example
```go
p := seo.New()
_ = p.Build([]byte(`{"title":"rfw","pattern":"%s | rfw"}`))
p.Install(nil)
seo.SetTitle("Docs")
seo.SetMeta("description", "rfw documentation")
```

## Notes and Limitations
Creates missing `<title>` and `<meta>` elements in the document `<head>`. The `pattern` uses `fmt.Sprintf` semantics and should include a single `%s` placeholder. `SetTitle` and `SetMeta` may be reassigned before `core.RegisterPlugin` to customize behavior.

## Related links
- [plugins](plugins)
- [dom](../dom)
