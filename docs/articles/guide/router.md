# Routing

The **router** maps paths to components and keeps navigation in sync with browser history. It resolves both `/docs` and `/docs/` to the same component, so trailing slashes don’t matter.

---

## Registering Routes

Define routes by pairing a path with a component:

```go
router.RegisterRoute(router.Route{
    Path: "/",
    Component: func() core.Component { return components.NewHome() },
})
```

---

## Handling Not Found Pages

Assign a component to `NotFoundComponent` for unmatched paths:

```go
router.NotFoundComponent = func() core.Component {
    return components.NewNotFoundComponent()
}
```

---

## When to Use

* Single-page applications with multiple views
* Navigating between dynamic content or user-generated routes
* Providing a fallback 404 component

## When Not to Use

* Very simple demos with a single view
* Cases where server-side routing already handles all navigation

---

## Demo

@include\:ExampleFrame:{code:"/examples/components/another\_component.go", uri:"/examples/user/jane"}

---

## Related

* [Router API](../api/router)
* [Pages Plugin](../plugins/pages)
