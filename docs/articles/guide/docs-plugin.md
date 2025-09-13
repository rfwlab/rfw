# docs plugin

The `docs` plugin powers the documentation site. It fetches the sidebar and markdown files on demand and emits browser events so components can render them.

| Event | Description |
| --- | --- |
| `rfwSidebar` | Fired after the sidebar HTML is loaded. |
| `rfwDoc` | Fired when a markdown document has been fetched. The event detail contains `path`, `content` and `headings`. |

## Loading articles

`LoadArticle` wraps the plugin's internal loader so components can request a markdown file.

```go
import docplug "github.com/rfwlab/rfw/v1/plugins/docs"

docplug.LoadArticle("/articles/guide.md")
```

### API

| Function | Description |
| --- | --- |
| `LoadArticle(path string)` | Fetches the markdown file at `path` and emits `rfwDoc` on completion. |

### Notes

- Requires the docs plugin to be installed.
- `path` must point to a file under `/articles/`.

## Meta tags

### Context
Sets the document `<title>` and `description` meta tag from each article's entry in `sidebar.json`.

### Prerequisites
- The `seo` plugin must be registered. It is enabled by default by the `docs` plugin.
- `sidebar.json` includes `title` and `description` for the article.

### How
1. Add the article to `sidebar.json` with `title` and `description` fields.
2. Register the plugin:
```go
core.RegisterPlugin(docs.New("/articles/sidebar.json"))
```
3. Load an article with `LoadArticle`. The metadata in `sidebar.json` becomes the page title and description.

### API
- `func SetTitle(title string)`
- `func SetMeta(name, content string)`

### Example
```go
docplug.LoadArticle("/articles/guide.md")
```

### Notes
- Disable SEO integration with `docs.New(path, true)`.
- If `description` is empty, the meta tag is cleared.

### Related
- [seo plugin](../api/seo)
- [js package](js)
- [docs plugin events](#events)
