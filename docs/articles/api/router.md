# router

Client-side router with lazy loaded components and guards.

- `RegisterRoute(Route)` adds a route definition.
- `Navigate(path)` programmatically changes the URL.
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

Routing parameters and query strings are merged and handled as in the following component.

@include:ExampleFrame:{code:"/examples/components/another_component.go", uri:"/examples/user/jane"}
