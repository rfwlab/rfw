# Docs Plugin

The **Docs plugin** powers documentation sites built with rfw. It loads sidebar and markdown files on demand, and emits browser events so components can react and render content.

## Features

* Load markdown articles dynamically.
* Emit browser events when sidebar or content is loaded.
* Integrates with the **SEO plugin** to update `<title>` and `<meta>` tags.

## Events

| Event        | Description                                                                            |
| ------------ | -------------------------------------------------------------------------------------- |
| `rfwSidebar` | Fired after the sidebar HTML is loaded.                                                |
| `rfwDoc`     | Fired when a markdown document is fetched. Includes `path`, `content`, and `headings`. |

## Usage

Register the plugin with a path to `sidebar.json`:

```go
import (
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/plugins/docs"
)

func main() {
    core.RegisterPlugin(docs.New("/articles/sidebar.json"))
}
```

Load an article:

```go
import docplug "github.com/rfwlab/rfw/v1/plugins/docs"

docplug.LoadArticle("/articles/guide.md")
```

## API Reference

| Function                        | Description                                               |
| ------------------------------- | --------------------------------------------------------- |
| `docs.LoadArticle(path string)` | Fetches a markdown file and emits `rfwDoc` on completion. |

## SEO Integration

* Requires the **SEO plugin** (enabled by default by `docs`).
* Reads `title` and `description` from `sidebar.json` and applies them as meta tags.

### Example Setup

1. Add article entry in `sidebar.json` with `title` and `description`.
2. Register the plugin: `core.RegisterPlugin(docs.New("/articles/sidebar.json"))`.
3. Load an article: `docs.LoadArticle("/articles/guide.md")`. Metadata is applied automatically.

### Notes

* Disable SEO integration with `docs.New(path, true)`.
* If `description` is empty, the meta tag will be cleared.
