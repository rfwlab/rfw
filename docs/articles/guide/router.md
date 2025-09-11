# Routing

## Why
Routing maps paths to components and keeps navigation in sync with browser history. The [Router API](../api/router) exposes registration and guard hooks.

```go
router.RegisterRoute(router.Route{
    Path: "/",
    Component: func() core.Component { return components.NewHome() },
})
```

## When to Use
Use the router to handle in-app navigation and not-found pages.

```go
router.NotFoundComponent = func() core.Component { return components.NewNotFoundComponent() }
```

## When Not to Use
Skip the router in single-view demos or when server-side routing already handles every path.

```go
// serve a static page without client-side navigation
```

## Interactive Demo
@include:ExampleFrame:{code:"/examples/components/another_component.go", uri:"/examples/user/jane"}
