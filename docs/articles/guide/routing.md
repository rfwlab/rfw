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

## Guards

`Route.Guards` holds the original guards: a `func(map[string]string) bool` that
allows or blocks. They keep working exactly as before, including
`router.Page(path, component, guards...)` and the `Group` builder.

`Route.ResultGuards` adds guards that say what should happen instead of the
route they refuse:

```go
router.RegisterRoute(router.Route{
    Path:      "/account",
    Component: NewAccountPage,
    ResultGuards: []router.ResultGuard{
        func(params map[string]string) router.GuardResult {
            switch {
            case !session.Authenticated():
                return router.ReplaceWith("/login")
            case !session.Can("account:read"):
                return router.Forbid()
            default:
                return router.Allow()
            }
        },
    },
})
```

- `router.Allow()` continues to the next guard and then to the route.
- `router.Forbid()` stops navigation with `router.ErrNavigationForbidden`.
- `router.RedirectTo(path)` navigates elsewhere and keeps a history entry for
  the current page. On an initial load or a browser back/forward, where the
  denied path is already the current entry, it replaces that entry instead, so
  back does not return to the page the guard refused.
- `router.ReplaceWith(path)` navigates elsewhere in place of the current entry,
  which is what a login redirect wants: back must not return to the page the
  user was denied.
- `router.HostReplace(path)` leaves the application and lets the host serve the
  path as a document.

`RedirectTo` and `ReplaceWith` are SPA navigation: the destination goes through
the same route matcher as `router.Navigate`, so it must be a registered route,
and an unregistered one lands on the not-found handling. Neither ever leaves
the application.

Guards run parent before child, and for each route its `Guards` run before its
`ResultGuards`. The first guard that does not allow decides: the guards behind
it, the route loader, the component factory, the unmount of the current page,
the history entry, and the DOM commit are all skipped. A redirect result
navigates to its destination without ever loading the protected route, bounded
by the same redirect-loop limit as `Route.Redirect`. A destination that is
empty or not a rooted path, and an unknown action, fail closed with
`router.ErrInvalidGuardResult`.

Do not call `router.Navigate` or `router.Replace` inside a guard: it reenters
navigation while one is in flight. Return the destination instead and let the
router apply it once. Outside browser builds the router keeps no history, so
`RedirectTo` and `ReplaceWith` both navigate to the destination.

### Handing a route to the host

Some pages belong to the server, not the SPA: a login form, a session-expiry
page, an SSO endpoint. `router.HostReplace(path)` hands the document to one:

```go
ResultGuards: []router.ResultGuard{
    func(map[string]string) router.GuardResult {
        if !session.Valid() {
            return router.HostReplace("/login?next=%2Fadmin")
        }
        return router.Allow()
    },
},
```

In the browser the guard chain ends there and the browser performs a real
navigation with `location.replace`. Nothing about the refused route is
committed first: no loader, no component, no unmount, no history entry, no DOM
write, and no route state change. The page the user is on stays exactly as it
is until the browser unloads it. Because the load replaces the current history
entry, back returns to the page before the protected one rather than to the
path the guard refused.

The target must already be a rooted path on the same origin, and it is validated
before any handoff, exactly as the guard returned it: nothing is trimmed or
otherwise repaired first, so a string that is not a valid target is refused
instead of being turned into one. Anything else fails closed with
`router.ErrInvalidGuardResult` and nothing reaches the browser: another scheme,
a scheme-relative `//host`, a backslash authority such as `/\host`, any ASCII
control character (browsers strip tabs and newlines from a URL, which turns
`/<tab>/host` into `//host`), a leading or trailing space, and a malformed
percent escape. A rooted path with a query and a fragment is accepted and is
handed over byte for byte.

Cross-origin host handoff is intentionally unsupported: a guard cannot send the
document to another origin, and there is no variant of this API that would. A
guard stays a declaration of where navigation may go, with no side effect of its
own, so the router remains the only thing that acts on the decision.

Outside browser builds there is no document to replace and the host page is not
a route the router can load, so `router.NavigateContext` returns a
`*router.HostNavigationError` carrying the target and wrapping
`router.ErrHostNavigation`. The current component, the active path and route
data are left untouched, and navigation status goes to `error`:

```go
var host *router.HostNavigationError
if errors.As(err, &host) {
    // host.Path is the path the browser build would have loaded
}
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
