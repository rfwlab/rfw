# router

```go
import "github.com/rfwlab/rfw/v2/router"
```

Client-side router with lazy loaded components, guards, and nested routes.

---

## Route Registration

### Page

```go
func Page(path string, component any, guards ...Guard)
```

Registers a route. `component` accepts:
- `func() *View` - called each navigation (fresh instance)
- `func() core.Component` - called each navigation (legacy)
- `*View` via `Singleton()` - reused every navigation

```go
router.Page("/", func() *types.View { return composition.NewFrom[Home]() })
router.Page("/admin", func() *types.View { return composition.NewFrom[Admin]() }, authGuard)
```

### Group

```go
func Group(prefix string, fn func(*GroupBuilder))
```

Creates nested routes under a common prefix. `GroupBuilder.Page` registers children:

```go
router.Group("/settings", func(g *router.GroupBuilder) {
    g.Page("profile", func() *types.View { return composition.NewFrom[Profile]() })
    g.Page("security", func() *types.View { return composition.NewFrom[Security]() })
})
```

### RegisterRoute

```go
func RegisterRoute(r Route)
```

Low-level registration. `Page` and `Group` call this internally.

---

## Singleton

```go
func Singleton(v *types.View) any
```

Wraps a pre-created `*View` so the same instance is reused on every navigation to that route. Pass the result as `Route.Component`:

```go
view := composition.New(&Dashboard{})
router.Page("/dashboard", router.Singleton(view))
```

---

## Route Struct

```go
type Route struct {
    Path      string
    Component any       // *View, func() *View, or func() core.Component
    Guards    []Guard
    Children  []Route
}
```

`Component` forms:
- `*View` - singleton, reused
- `func() *View` - factory, called each navigation
- `func() core.Component` - factory (legacy)

`Guard` is `func(map[string]string) bool`. Return `false` to block navigation.

---

## Navigation

### Navigate

```go
func Navigate(fullPath string)
```

Programmatically navigates. Supports query strings (`/search?q=test`). Triggers unmount/mount lifecycle. Calls guards before proceeding.

### CanNavigate

```go
func CanNavigate(fullPath string) bool
```

Reports whether `fullPath` matches a registered route.

### ExposeNavigate

```go
func ExposeNavigate()
```

Exposes `goNavigate(path)` to JavaScript and auto-intercepts internal `<a href>` clicks that match registered routes.

### InitRouter

```go
func InitRouter()
```

Starts the router. Listens for `popstate` events and navigates to the current URL.

---

## Inspection

### RegisteredRoutes

```go
func RegisteredRoutes() []RegisteredRoute
```

Returns all registered routes with resolved full paths, dynamic params, and nested children. Useful for tooling and diagnostics.

```go
type RegisteredRoute struct {
    Template string
    Path     string
    Params   []string
    Children []RegisteredRoute
}
```

### CurrentComponent

```go
func CurrentComponent() core.Component
```

Returns the currently mounted routed component.

---

## Not Found

```go
var NotFoundComponent any           // component or factory for unmatched routes
var NotFoundCallback func(string)   // called with the unmatched path
```

If `NotFoundCallback` is set, it takes priority. Otherwise `NotFoundComponent` is rendered.

---

## Reset

```go
func Reset()
```

Clears all routes and the current component. Intended for tests.