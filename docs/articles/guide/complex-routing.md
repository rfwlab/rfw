# Complex Routing

Routing in **rfw** supports nested paths and multiple dynamic parameters. This allows you to build deep links and hierarchies.

## Example

Register a route with two dynamic parameters:

```go
router.RegisterRoute(router.Route{
    Path: "/complex/:user/:section",
    Component: func() core.Component { return components.NewComplexRoutingComponent() },
})
```

Navigating to `/complex/alice/settings` renders the component with the route parameters injected as props:

* `user = "alice"`
* `section = "settings"`

This makes it easy to bind URL segments directly to component data.

## Interactive Example

@include\:ExampleFrame:{code:"/examples/components/complex\_routing\_component.go", uri:"/examples/complex/jane/profile"}

## Use Cases

* Nested account pages like `/users/:id/settings`
* Project hierarchies such as `/projects/:id/tasks/:taskId`
* Any scenario where multiple parameters need to be parsed from the path
