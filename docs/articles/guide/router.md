# Routing

The router maps paths to components and supports nested routes and navigation guards:

```go
router.RegisterRoute(router.Route{
    Path: "/",
    Component: func() core.Component { return components.NewHome() },
})
```

Guards can block navigation by returning `false`.
Client navigation uses declarative routes.

@include:ExampleFrame:{code:"/examples/components/another_component.go", uri:"/examples/user/jane"}
