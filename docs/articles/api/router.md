# router

Client-side router with lazy loaded components and guards.

- `RegisterRoute(Route)` adds a route definition.
- `Navigate(path)` programmatically changes the URL.
- `InitRouter()` starts the router and listens for navigation events.
- `ExposeNavigate()` exposes navigation to JavaScript as `goNavigate`.
- `NotFoundComponent` or `NotFoundCallback` handle unmatched routes.
- `Children []Route` nests routes under a parent.
- Guards: `Guard` functions run before navigation and can cancel by returning `false`.

## Usage

Routes are defined with `router.RegisterRoute` by specifying the path and the
component to mount. `router.InitRouter` starts the router.

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

Call `router.ExposeNavigate()` to register a global `goNavigate(path)` function.
This is often combined with `router.InitRouter()` to allow links or other
handlers to navigate without reloading the page.

```go
router.ExposeNavigate()
router.InitRouter()
```

```html
<a href="/docs" onclick="goNavigate('/docs'); return false">Docs</a>
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
