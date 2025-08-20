# router

Client-side router with lazy loaded components and guards.

- `RegisterRoute(Route)` adds a route definition.
- `Navigate(path)` programmatically changes the URL.
- Guards: `Guard` functions run before navigation and can cancel by returning `false`.

## Usage

Routes are defined with `router.RegisterRoute` by specifying the path and the
component to mount. `router.InitRouter` starts the router.

## Example

```go
router.RegisterRoute(router.Route{
        Path: "/",
        Component: func() core.Component { return components.NewMainComponent() },
})
router.InitRouter()
```

1. `RegisterRoute` associates the root path with the main component.
2. `InitRouter` listens for URL changes and renders the current route.
