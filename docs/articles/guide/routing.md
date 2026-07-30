# Routing

Routes may carry a stable name, metadata, a redirect, or an asynchronous data
loader.

## Names and URLs

```go
router.RegisterRoute(router.Route{
    Path: "/teams/:team",
    Children: []router.Route{{
        Path:      "users/:user",
        Name:      "team-user",
        Component: NewUserPage,
        Meta:      map[string]any{"section": "users"},
    }},
})

path := router.MustURL("team-user", map[string]string{
    "team": "core",
    "user": "ada",
}, url.Values{"tab": {"activity"}})
```

`URL` returns an error for an unknown route name or a missing path parameter.
Values are path escaped and query values use `url.Values.Encode`.

Redirect routes do not need a component:

```go
router.RegisterRoute(router.Route{
    Path:     "/people/:id",
    Redirect: "/users/:id",
})
```

## Data loaders

A loader completes before the new component is created and the previous page
is unmounted:

```go
router.RegisterRoute(router.Route{
    Path:      "/users/:id",
    Component: NewUserPage,
    Loader: func(ctx context.Context, load router.LoadContext) (any, error) {
        return api.User(ctx, load.Params["id"])
    },
})
```

The component may receive the result by implementing:

```go
func (page *UserPage) SetRouteData(data any) {
    page.user = data.(User)
}
```

Starting another navigation cancels the current loader. A failed or cancelled
loader leaves the mounted page in place. `router.NavigateContext` waits for
the result and returns its error, which is useful in tests and controlled
workflows.

The following signals expose transition state:

```go
router.Status() // idle, loading, ready, error
router.Error()
router.Data()
router.Meta()
router.ActivePath()
```

Browser history navigation restores saved scroll positions. New and replaced
entries start at the top. Set `Meta["preserveScroll"]` to `true` for a route
that owns its scroll behavior, or call `router.SetScrollRestoration(false)` to
disable router management.

## Built-in component wrappers

- `core.NewPortal(selector, child)` mounts a child below another DOM target.
- `core.NewKeepAlive(child)` detaches and caches a child's DOM on route exit.
  Call `Dispose` when the cache should be released permanently.
- `core.NewTransition(child, config)` applies CSS enter and leave classes.
  Leave content is moved outside the router outlet until the configured
  duration ends.

A `KeepAlive` instance should be registered directly as the route component
so the router treats it as a singleton:

```go
router.Page("/editor", core.NewKeepAlive(NewEditor()))
```
