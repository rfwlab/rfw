# Docs Plugin

The **docs plugin** powers the documentation site. It loads sidebar and markdown files on demand and emits browser events so components can render them.

## Events

| Event        | Description                                                                                              |
| ------------ | -------------------------------------------------------------------------------------------------------- |
| `rfwSidebar` | Fired after the sidebar HTML is loaded                                                                   |
| `rfwDoc`     | Fired when a markdown document has been fetched; event detail includes `path`, `content`, and `headings` |

## Loading Articles

Use `LoadArticle` to request a markdown file:

```go
import docplug "github.com/rfwlab/rfw/v1/plugins/docs"

docplug.LoadArticle("/articles/guide.md")
```

### API

| Function                   | Description                                              |
| -------------------------- | -------------------------------------------------------- |
| `LoadArticle(path string)` | Fetches a markdown file and emits `rfwDoc` on completion |

**Notes**

* Requires the docs plugin
* `path` must point under `/articles/`

## Meta Tags

The plugin integrates with the **seo plugin** to set `<title>` and `<meta name="description">` from entries in `sidebar.json`.

### Requirements

* `seo` plugin must be registered (enabled by default by `docs`)
* `sidebar.json` must include `title` and `description`

### Setup

1. Add the article to `sidebar.json` with `title` and `description`
2. Register the plugin:

   ```go
   core.RegisterPlugin(docs.New("/articles/sidebar.json"))
   ```
3. Load an article with `LoadArticle`. Metadata is applied automatically.

### API

* `var SetTitle func(title string)`
* `var SetMeta func(name, content string)`

### Example

```go
docplug.LoadArticle("/articles/guide.md")
```

### Notes

* Disable SEO integration with `docs.New(path, true)`
* If `description` is empty, the meta tag is cleared

## Related

* [seo plugin](../api/seo)
* [js package](js)
* [docs plugin events](#events)
