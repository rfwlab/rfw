# Routing

The **router** maps URL paths to components and keeps navigation in sync with browser history. Both `/docs` and `/docs/` resolve to the same component, trailing slashes don't matter.

---

## Registering Routes

### `router.Page()`

The preferred API. Registers a single route with a path, component, and optional guards:

```go
router.Page("/", func() *types.View {
    return composition.New(&Home{})
})
```

`router.Page` accepts the same component forms as `Route.Component` (see below).

### `router.Group()`

Creates nested routes under a common path prefix:

```go
router.Group("/admin", func(r *router.GroupBuilder) {
    r.Page("/dashboard", func() *types.View {
        return composition.New(&Dashboard{})
    })
    r.Page("/settings", func() *types.View {
        return composition.New(&Settings{})
    })
})
```

This registers `/admin/dashboard` and `/admin/settings`.

### `router.Singleton()`

Wraps a pre-created `*types.View` so every navigation reuses the same instance:

```go
home := composition.New(&Home{})
router.Page("/", router.Singleton(home))
```

Use singletons when a component is stateless or you want to preserve state across navigations.

### `router.RegisterRoute()`

The low-level API, still available but `Page()` is preferred:

```go
router.RegisterRoute(router.Route{
    Path:      "/",
    Component: func() *types.View { return composition.New(&Home{}) },
})
```

---

## Component Forms

`Route.Component` (and the second argument to `Page()`) accepts three forms:

| Form | Behavior |
|------|----------|
| `func() *types.View` | Called each navigation, fresh instance every time |
| `*types.View` | Singleton, reused every navigation |
| `func() core.Component` | Called each navigation (legacy) |

```go
// Fresh instance each navigation
router.Page("/items", func() *types.View {
    return composition.New(&ItemList{})
})

// Singleton, same instance reused
view := composition.New(&Layout{})
router.Page("/", router.Singleton(view))
```

---

## Path Parameters

Use `:name` segments to capture dynamic parts of the path:

```go
router.Page("/users/:id", func() *types.View {
    return composition.New(&UserProfile{})
})
```

When navigating to `/users/42`, the component receives `map[string]string{"id": "42"}` via `SetRouteParams`. Access params in `OnMount`:

```go
type UserProfile struct {
    composition.Component
}

func (u *UserProfile) OnMount() {
    params := u.HTMLComponent.RouteParams()
    id := params["id"]
    // fetch user by id...
}
```

Query parameters are merged into the same params map.

---

## Guards

Guards control whether navigation is allowed. Pass them as variadic arguments to `Page()`:

```go
func requireAuth(params map[string]string) bool {
    return session.IsAuthenticated()
}

router.Page("/dashboard", func() *types.View {
    return composition.New(&Dashboard{})
}, requireAuth)
```

If any guard returns `false`, navigation is blocked. If the current component is `nil`, the router falls back to `/`.

---

## Not Found

Set `NotFoundComponent` for unmatched paths:

```go
router.NotFoundComponent = func() *types.View {
    return composition.New(&NotFound{})
}
```

Accepts the same component forms as `Route.Component`.

---

## Programmatic Navigation

### `router.Navigate()`

```go
router.Navigate("/users/42")
```

Renders the matching component, pushes browser history, and passes path/query params.

### `router.ExposeNavigate()`

Makes `Navigate` available from JavaScript and intercepts internal `<a>` clicks:

```go
router.ExposeNavigate()
```

After calling `ExposeNavigate`, clicking `<a href="/about">` triggers client-side navigation instead of a full page load, provided the path matches a registered route.

---

## `router.CanNavigate()`

Reports whether a path matches a registered route, useful for pre-checking before navigation:

```go
if router.CanNavigate("/secret") {
    router.Navigate("/secret")
}
```

---

## `router.InitRouter()`

Call once at startup to begin listening for `popstate` events and navigate to the current URL:

```go
func main() {
    router.Page("/", func() *types.View {
        return composition.New(&Home{})
    })
    router.InitRouter()
    select {}
}
```

---

## Full Example

```go
//go:build js && wasm

package main

import (
    "embed"

    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/router"
    "github.com/rfwlab/rfw/v2/types"
)

//go:embed Home.rtml About.rtml
var fs embed.FS

func init() {
    composition.RegisterFS(&fs)
}

type Home struct{ composition.Component }
type About struct{ composition.Component }

func requireAuth(params map[string]string) bool {
    return true
}

func main() {
    router.Page("/", func() *types.View {
        return composition.New(&Home{})
    })

    router.Group("/app", func(r *router.GroupBuilder) {
        r.Page("/about", func() *types.View {
            return composition.New(&About{})
        }, requireAuth)
    })

    router.NotFoundComponent = func() *types.View {
        return composition.New(&Home{})
    }

    router.ExposeNavigate()
    router.InitRouter()
    select {}
}
```

---

## Related

- [Router API](../api/router)
- [Signals and Stores](/docs/guide/store-vs-signals)