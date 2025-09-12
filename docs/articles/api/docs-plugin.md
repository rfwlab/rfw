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

### Related

- [js package](js.md)
- [docs plugin events](#events)
