# Creating an Application

This guide walks through setting up an **rfw v2** project, project layout, template embedding, root component, routing, and SSC configuration.

---

## Project Layout

```
myapp/
├── main.go
├── components/
│   ├── home.go
│   ├── counter.go
│   └── templates/
│       ├── Home.rtml
│       └── Counter.rtml
├── static/
│   └── (images, styles, etc.)
└── rfw.json
```

- `components/`, Go structs and the `composition` API. Templates live in a `templates/` subdirectory (any depth works with `go:embed`).
- `static/`, assets copied into `build/static/` and served at `/`.
- `rfw.json`, build and runtime configuration.

---

## Embedding Templates

In v2, templates are resolved from `embed.FS` instances registered at init time. Place your `.rtml` files under a `templates` directory and use `//go:embed`:

```go
package components

import (
    _ "embed"
    "embed"

    "github.com/rfwlab/rfw/v2/composition"
)

//go:embed templates
var templates embed.FS

func init() {
    composition.RegisterFS(&templates)
}
```

`composition.RegisterFS` adds the FS to a global resolver. When `composition.New` looks for a template, it searches all registered FS instances. Call `RegisterFS` in `init()` or at package level, it must run before any `composition.New` calls.

---

## Main Entry Point

The entry point registers routes and starts the router:

```go
package main

import (
    "github.com/rfwlab/rfw/v2/router"
    _ "github.com/username/myapp/components"
)

func main() {
    homeView, _ := composition.NewFrom[components.Home]()
    counterView, _ := composition.NewFrom[components.Counter]()
    router.Page("/", homeView)
    router.Page("/counter", counterView)
    router.InitRouter()
    select {}
}
```

### `router.Page`

`router.Page(path, component)` is the v2 shorthand for registering a route. It accepts a `*View`, a `func() *View`, or a `func() core.Component`:

```go
view, err := composition.New(&About{})
if err != nil {
    log.Fatal(err)
}
router.Page("/", view)                        // singleton, same instance every visit
router.Page("/about", func() *composition.View {
    v, _ := composition.New(&About{})
    return v
})
```

### `router.Group`

For nested routes under a common prefix:

```go
profileView, _ := composition.NewFrom[Profile]()
securityView, _ := composition.NewFrom[Security]()
router.Group("/settings", func(g *router.GroupBuilder) {
    g.Page("/profile", profileView)
    g.Page("/security", securityView)
})
```

This registers `/settings/profile` and `/settings/security`.

---

## Root Component

Components embed `composition.Component` and use typed fields for type-based auto-wiring. The template is found automatically by struct name:

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/types"
)

type Home struct {
    composition.Component
    Count   *t.Int
    Body    *t.View
}

func (h *Home) OnMount() {
    h.Count.Set(0)
}

func (h *Home) Increment() {
    h.Count.Set(h.Count.Get() + 1)
}
```

The struct name `Home` resolves to `Home.rtml` (or `Home.html`) in any registered `embed.FS`. Override this by implementing the `Template() string` method:

```go
func (h *Home) Template() string {
    return "pages/index.rtml"
}
```

The corresponding template:

```rtml
<root>
  <h1>Home</h1>
  <p>@signal:Count</p>
  <button @on:click:Increment>+</button>
  @include:content
</root>
```

### Wiring Summary

| Field Type | Effect |
|-----------|--------|
| `*t.Int`, `*t.String`, `*t.Bool`, `*t.Float`, `*t.Any` | Detected as reactive signals; nil fields are auto-initialized |
| `*t.Store` | Creates a component-scoped store by field name |
| `*t.Inject[T]` | Resolves type T from the DI container |
| `*t.View` | Slot for composition (slot name = lowercase field name) |
| `*t.Ref` | Template ref for DOM element access |
| `*t.History` | Wires undo/redo handlers |
| `t.HInt`, `t.HString`, etc. | Host signal types for server-side data |
| `t.Prop[T]` | Declared prop of type T |
| `Template() string` method | Overrides convention-based template resolution |

---

## SSC Mode

rfw v2 uses **Server-Side Composition (SSC)**, no SPA fallback. Configure it in `rfw.json`:

```json
{
  "mode": "ssc",
  "entry": "main.go",
  "output": "build"
}
```

With SSC, every navigation is a real page load processed by the server. The router handles client-side transitions for registered routes, but there is no client-side fallback for unmatched paths, the server must serve the correct HTML shell.

---

## Next Steps

- [Components Basics](./components-basics), typed fields, signals, includes, and lifecycle
- [Composition](./composition), how `composition.New` works under the hood
- [Template Syntax](./template-syntax), RTML directives and expressions