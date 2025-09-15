# Pages Plugin

The **Pages plugin** provides **file‑based routing**. Drop Go files into a `pages/` folder and get routes automatically—no manual router wiring.

> Each file becomes a route, folders become path segments, and `[param]` in a filename becomes a dynamic parameter.

## What it does

* Scans `pages/` during **PreBuild** and generates a temporary `routes_gen.go` with `router.RegisterRoute(...)` calls.
* Removes the generated file after the build.
* On `rfw dev`, rescans so new pages appear instantly.

Activate registrations with a blank import in `main.go`:

```go
import _ "your/module/pages"
```

## Mapping rules

| Filesystem                   | Route path     | Constructor                             |
| ---------------------------- | -------------- | --------------------------------------- |
| `pages/index.go`             | `/`            | `func Index() core.Component`           |
| `pages/about.go`             | `/about`       | `func About() core.Component`           |
| `pages/posts/[id].go`        | `/posts/:id`   | `func PostsId() core.Component`         |
| `pages/admin/users/index.go` | `/admin/users` | `func AdminUsersIndex() core.Component` |

**Dynamic params** are **injected as props**: use `{id}` in templates; read from the component props in Go if needed.

## Required export

Each page must export a **PascalCase** constructor derived from its path. Keep constructors argument‑less—the router provides params via props.

```
pages/
  index.go        // -> func Index() core.Component
  about.go        // -> func About() core.Component
  posts/[id].go   // -> func PostsId() core.Component
```

## Example

**`pages/posts/[id].go`**

```go
package posts

import (
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/composition"
)

//go:embed [id].rtml
var tpl []byte

func PostsId() core.Component {
    cmp := composition.Wrap(core.NewComponent("PostPage", tpl, nil))
    return cmp.HTML()
}
```

**`pages/posts/[id].rtml`**

```rtml
<root>
  <h1>Post #{id}</h1>
  <p>Loading post content…</p>
</root>
```

Navigating to `/posts/42` renders the page with `{id}` = `42`.

## Configuration

Enable the plugin in `rfw.json` (you can customize the scanned directory via `dir`):

```json
{
  "plugins": {
    "pages": {
      "dir": "pages"
    }
  }
}
```

* **dir** — source directory that contains page files (default: `pages`).

## Best practices

* **Keep pages thin**: layout & data loading in pages, reusable UI in `components/`.
* **Name for URLs**: filenames/folders become paths—prefer `kebab-case` for readability.
* **Use dynamic segments sparingly**: `[id].go` for detail pages; for runtime‑driven routes, register them programmatically.
* **Do not commit** `routes_gen.go` (it’s temporary).

### Do / Don’t

* ✅ Use `index.go` for folder roots (`/`, `/admin`, …)
* ✅ Colocate templates with page code (e.g. `[id].rtml`)
* ❌ Don’t pass constructor args—use route params via props instead

## When to use

Choose **file‑based routing** for fast iteration and URL‑driven structure (dashboards, docs, content‑heavy sites).

Prefer **manual router registration** when you need dynamic routes built at runtime, guards/redirects with fine‑grained control, or nonstandard resolution (e.g., experiments that swap components).
