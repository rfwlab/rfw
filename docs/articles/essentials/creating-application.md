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
    router.Page("/", composition.NewFrom[components.Home]())
    router.Page("/counter", composition.NewFrom[components.Counter]())
    router.InitRouter()
    select {}
}
```

### `router.Page`

`router.Page(path, component)` is the v2 shorthand for registering a route. It accepts a `*View`, a `func() *View`, or a `func() core.Component`:

```go
router.Page("/", view)                        // singleton, same instance every visit
router.Page("/about", func() *composition.View {
    return composition.New(&About{})
})
```

### `router.Group`

For nested routes under a common prefix:

```go
router.Group("/settings", func(g *router.GroupBuilder) {
    g.Page("/profile", composition.NewFrom[Profile]())
    g.Page("/security", composition.NewFrom[Security]())
})
```

This registers `/settings/profile` and `/settings/security`.

---

## Root Component

Components embed `composition.Component` and use struct tags for wiring. The template is found automatically by struct name:

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/types"
)

type Home struct {
    composition.Component
    Count *composition.Int `rfw:"signal"`
    Body  *types.View     `rfw:"include:content"`
}

func (h *Home) OnMount() {
    h.Count.Set(0)
}

func (h *Home) Increment() {
    h.Count.Set(h.Count.Get() + 1)
}
```

The struct name `Home` resolves to `Home.rtml` (or `Home.html`) in any registered `embed.FS`. Override this with a `rfw:"template:path"` tag on a blank field:

```go
type Home struct {
    composition.Component
    _    struct{}        `rfw:"template:pages/index.rtml"`
    Count *composition.Int `rfw:"signal"`
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

| Tag | Field Type | Effect |
|-----|-----------|--------|
| `rfw:"signal"` | `*composition.Int`, `*composition.String`, etc. | Registers as a reactive prop; nil fields are auto-initialized |
| `rfw:"store"` | `*composition.Store` | Creates a component-scoped store by field name |
| `rfw:"inject"` | any | Resolves from the DI container; defaults to field name |
| `rfw:"include:slotname"` | `*types.View` | Calls `AddDependency(slotname, view)` for composition |
| `rfw:"template:path"` | blank struct field | Overrides convention-based template resolution |
| `rfw:"host"` | string | Registers a host component by name |
| `rfw:"history:store:undo:redo"` |, | Wires undo/redo handlers onto the named store |

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

- [Components Basics](./components-basics), struct tags, signals, includes, and lifecycle
- [Composition](./composition), how `composition.New` works under the hood
- [Template Syntax](./template-syntax), RTML directives and expressions