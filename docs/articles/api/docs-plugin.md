# docs plugin

The `docs` plugin powers the documentation site. It fetches the sidebar and markdown files on demand and emits browser events so components can render them.

## Events

The plugin dispatches these custom events on `document`:

- `rfwSidebar` after the sidebar HTML is loaded.
- `rfwDoc` when a markdown document has been fetched. The event detail contains:
  - `path` – the document path
  - `content` – raw markdown source
  - `headings` – extracted heading objects with `text`, `depth` and `id`

These headings enable the interface to render a table of contents for each page.
