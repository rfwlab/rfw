# router

Client-side router with lazy loaded components and guards.

- `RegisterRoute(Route)` adds a route definition.
- `Navigate(path)` programmatically changes the URL.
- `InitRouter()` starts the router and listens for navigation events.
- `ExposeNavigate()` exposes navigation to JavaScript as `goNavigate`.
- `NotFoundComponent` or `NotFoundCallback` handle unmatched routes.
- Guards: `Guard` functions run before navigation and can cancel by returning `false`.

## Usage

Routes are defined with `router.RegisterRoute` by specifying the path and the
component to mount. `router.InitRouter` starts the router.

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

During navigation, the router merges path parameters and query string values and calls `SetRouteParams` before the component mounts. The following component demonstrates handling parameters.

@include:ExampleFrame:{code:"/examples/components/another_component.go", uri:"/examples/user/jane"}
