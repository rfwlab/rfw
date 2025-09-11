# docs plugin

The `docs` plugin powers the documentation site. It fetches the sidebar and markdown files on demand and emits browser events so components can render them.

| Event | Description |
| --- | --- |
| `rfwSidebar` | Fired after the sidebar HTML is loaded. |
| `rfwDoc` | Fired when a markdown document has been fetched. The event detail contains `path`, `content` and `headings`. |

## Events

The plugin dispatches these custom events on `document`.
