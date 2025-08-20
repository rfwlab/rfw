# Complex Routing

Complex routing in **rfw** allows nested paths and multiple dynamic
parameters. The example application registers a route that captures a user
and section from the URL:

```go
router.RegisterRoute(router.Route{
    Path: "/complex/:user/:section",
    Component: func() core.Component { return components.NewComplexRoutingComponent() },
})
```

Navigating to `/complex/alice/settings` renders the component with the
route parameters injected as props, enabling deep links and hierarchies.
