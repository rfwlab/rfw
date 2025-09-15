# Pages Plugin

The **Pages** plugin gives you *file‑based routing* for rfw: drop Go files into a `pages/` folder and get routes automatically. It’s the fastest way to scaffold page‑level UI without wiring the router by hand.

> TL;DR: each file becomes a route, folders become path segments, and `[param]` in a filename becomes a dynamic parameter.

---

## Why use it

* **Zero boilerplate** – no manual `router.RegisterRoute` for every page.
* **Convention over configuration** – URLs mirror your folder structure.
* **Great defaults for content & apps** – perfect for marketing sites, docs, dashboards.
* **Still flexible** – you can mix file‑based pages with programmatic routes when needed.

---

## How it works

During **PreBuild**, the plugin scans the `pages/` directory and generates a temporary `routes_gen.go` with calls to `router.RegisterRoute(...)`. After the build, the generated file is removed. The same scanning happens on `rfw dev` so new files appear instantly.

To activate the registrations, import the generated package (blank import) in `main.go`:

```go
import _ "your/module/pages"
```

### Mapping rules

| Filesystem                   | Route path     | Constructor                             |
| ---------------------------- | -------------- | --------------------------------------- |
| `pages/index.go`             | `/`            | `func Index() core.Component`           |
| `pages/about.go`             | `/about`       | `func About() core.Component`           |
| `pages/posts/[id].go`        | `/posts/:id`   | `func PostsId() core.Component`         |
| `pages/admin/users/index.go` | `/admin/users` | `func AdminUsersIndex() core.Component` |

> Dynamic params are **injected as props**: reference them in templates with `{id}` and access them in Go via the component props map if needed.

### Required export

Each page file must export a constructor whose name is the **PascalCase** version of its path. For example:

```
pages/
  index.go        // -> func Index() core.Component
  about.go        // -> func About() core.Component
  posts/[id].go   // -> func PostsId() core.Component
```

> Keep constructors argument‑less. The router will provide route params as props.

---

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

---

## Configuration

Add the plugin to `rfw.json`. You can customise the scanned directory via `dir`.

```json
{
  "plugins": {
    "pages": {
      "dir": "pages"
    }
  }
}
```

* **dir**: source directory that contains page files (default: `pages`).

---

## Best practices

* **Keep pages thin**: focus page files on layout & data loading. Reusable pieces belong in `components/`.
* **Name for URLs**: folders and filenames become paths—prefer `kebab-case` for readability.
* **Use dynamic segments sparingly**: `[id].go` is good for detail pages; if routing depends on runtime data, register it programmatically.
* **Don’t commit generated code**: `routes_gen.go` is temporary.

### Do / Don’t

* ✅ Use `index.go` for folder roots (`/`, `/admin`, …)
* ✅ Colocate templates with page code (e.g. `[id].rtml`)
* ❌ Don’t pass constructor args—use route params via props instead

---

## When to use the Pages plugin

Choose file‑based routing when you want fast iteration and URL‑driven structure. For dashboards, docs, or content‑heavy sites, it’s usually the best default.

Use **manual router registration** when you need:

* dynamic routes built at runtime
* fine‑grained control over guards, redirects, or conditional mounts
* nonstandard resolution (e.g. A/B experiments that swap components)

---

## FAQ

**How do I read params in code?**  They are injected as props—reference them in templates with `{param}`. In Go, read from the component’s props map.

**Does it support nested routes?**  Yes—folder nesting becomes path nesting.

**What about 404s or catch‑alls?**  Not part of this plugin—register a fallback route with the router API.

---

## See also

* [Router basics](/docs/essentials/routing)
* [Components](/docs/essentials/components-basics)
