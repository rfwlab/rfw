# router

Client-side router with lazy loaded components and guards.

| Function | Description |
| --- | --- |
| `RegisterRoute(Route)` | Adds a route definition. |
| `Navigate(path)` | Programmatically changes the URL. |
| `CanNavigate(path) bool` | Reports whether a path matches a registered route. |
| `InitRouter()` | Starts the router and listens for navigation events. |
| `ExposeNavigate()` | Exposes navigation to JavaScript as `goNavigate` and auto-routes internal links. |
| `NotFoundComponent` / `NotFoundCallback` | Handle unmatched routes. |
| `Reset()` | Clears registered routes and the current component, useful in tests. |
| `Route.Children []Route` | Nests routes under a parent. |
| `Guard` | Runs before navigation and can cancel by returning `false`. |

`CanNavigate` helps determine whether a path is registered before navigating:

```go
if router.CanNavigate(path) {
    router.Navigate(path)
}
```

## Usage

Routes are defined with `router.RegisterRoute` by specifying the path and the
component to mount. `router.InitRouter` starts the router.

### Trailing slashes

Routes accept an optional trailing `/`. Registering `Path: "/trail"` matches
both `/trail` and `/trail/`, while `/trail/extra` is still treated as
unregistered.

```go
router.RegisterRoute(router.Route{
    Path: "/trail",
    Component: func() core.Component { return components.NewTrailComponent() },
})
```

### Nested routes

`Children []Route` lets a route own additional sub-routes. Child paths are
registered relative to the parent and share its component tree.

```go
router.RegisterRoute(router.Route{
    Path: "/parent",
    Component: func() core.Component { return components.NewParentComponent() },
    Children: []router.Route{
        {
            Path: "/parent/child",
            Component: func() core.Component { return components.NewChildComponent() },
        },
    },
})
```

### Navigating from JavaScript

Call `router.ExposeNavigate()` to register a global `goNavigate(path)` function
and automatically route internal `<a>` clicks. Combine it with
`router.InitRouter()` so the initial page loads correctly.

```go
router.ExposeNavigate()
router.InitRouter()
```

```html
<a href="/docs">Docs</a> <!-- internal links are routed -->
<button onclick="goNavigate('/docs')">Docs</button> <!-- manual trigger -->
```

Components that need access to route parameters can implement the `routeParamReceiver` interface:

```go
type routeParamReceiver interface {
    SetRouteParams(map[string]string)
}
```

During navigation, the router merges path parameters and query string values
before calling `SetRouteParams` on the target component.

```go
router.RegisterRoute(router.Route{
    Path: "/examples/params/:id",
    Component: func() core.Component { return components.NewParamsComponent() },
})
router.Navigate("/examples/params/42?tab=posts")
// ParamsComponent.SetRouteParams receives: map[string]string{"id": "42", "tab": "posts"}
```

@include:ExampleFrame:{code:"/examples/components/params_component.go", uri:"/examples/params/42?tab=posts"}

### Navigation guards

Guards run before navigation. If any guard returns `false`, the router aborts
the transition.

```go
router.RegisterRoute(router.Route{
    Path: "/protected",
    Component: func() core.Component { return components.NewProtectedComponent() },
    Guards: []router.Guard{
        func(map[string]string) bool { return false },
    },
})
router.Navigate("/protected") // navigation is cancelled
```
