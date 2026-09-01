# State, asynchronous data, and lifecycle

rfw signals can group writes, cache derived values, and represent cancellable
asynchronous work without a second state library.

## Batch and memo

`state.Batch` defers dependent effects until the callback finishes. Each
effect runs once even when several dependencies change:

```go
state.Batch(func() {
    firstName.Set("Ada")
    lastName.Set("Lovelace")
})
```

`state.Memo` tracks the signals read by its compute function and stores the
latest result:

```go
fullName := state.Memo(func() string {
    return firstName.Get() + " " + lastName.Get()
})
defer fullName.Stop()
```

Use `state.Untracked` when an effect needs a value without subscribing to it.

## Resources and Suspense

A resource starts its loader immediately, exposes reactive status, and
cancels pending work when closed:

```go
user := state.NewResource(func(ctx context.Context) (User, error) {
    return loadUser(ctx, "42")
},
    state.WithResourceKey("user:42"),
    state.WithResourceTTL(time.Minute),
)
defer user.Close()
```

Resources with the same non-empty key share an in-flight request and cached
value. `Reload` clears the key before loading, `Mutate` installs a local value,
and `Invalidate` returns the resource to idle.

`Read` returns `state.ErrResourcePending` until data is ready. `Suspense`
tracks the resource and patches its own root when the status changes:

```go
view := core.NewSuspense(func() (string, error) {
    value, err := user.Read()
    if err != nil {
        return "", err
    }
    return "<p>" + html.EscapeString(value.Name) + "</p>", nil
}, "<p>Loading...</p>")
```

## Component scope

Every `HTMLComponent` owns a `core.Scope`. The scope closes on unmount and is
created again on a later mount:

```go
component.Scope().Go(func(ctx context.Context) {
    pollUntilCancelled(ctx)
})
component.Scope().Defer(subscription.Stop)
component.Effect(func() func() {
    render(count.Get())
    return nil
})
```

`Scope.Go` receives the component context. `Scope.Defer` runs cleanup in
reverse registration order. Registering cleanup after the scope has closed
runs it immediately.

## DOM lifecycle hooks and browser libraries

`DOMHook` is the boundary for browser APIs and vendored JavaScript libraries:

```go
component.DOMHook(dom.LifecycleHook{
    Mounted: func(root dom.Element) func() {
        chart := js.Get("Chart").New(root.Query("canvas").Value, options)
        return func() {
            chart.Call("destroy")
        }
    },
    Updated: func(root dom.Element) {
        // Synchronize an existing browser widget after a DOM patch.
    },
})
```

This is not npm compatibility in the package-manager sense. rfw does not
resolve npm packages, run Node, or bundle JavaScript modules. A library that
publishes a browser-ready file can be copied into `static/`, loaded by the
page, and initialized through its global browser API. The lifecycle cleanup
keeps that integration tied to the component that owns it.
