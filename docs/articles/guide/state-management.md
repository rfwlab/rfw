# State Management

Reactivity in **rfw** is powered by **stores** and **signals**. Stores group values under a module and key, making them available across components. Signals represent local reactive values tied to a component. Together they provide a flexible and predictable way to manage application state.

## Stores

Create a store with `state.NewStore`:

```go
import "github.com/rfwlab/rfw/v1/state"

s := state.NewStore("profile", state.WithModule("user"))
s.Set("first", "Ada")
s.Set("last", "Lovelace")
```

* Namespaces: add `WithModule` to scope keys.
* Persistence: add `WithPersistence` to store values in localStorage.
* History: add `WithHistory` to enable undo/redo.

Stores automatically notify subscribed components when values change.

## Signals

Use signals for local, component-only state:

```go
count := state.NewSignal(0)
```

Signals are simple and efficient when you don’t need global access.

## Computed Values

Derive new values from existing keys:

```go
state.Map2(s, "fullName", "first", "last", func(first, last string) string {
    return first + " " + last
})
```

Or define custom computed values:

```go
s.RegisterComputed(state.NewComputed(
    "profile",
    []string{"first", "last", "age"},
    func(m map[string]any) any {
        return fmt.Sprintf("%s %s (%d)", m["first"], m["last"], m["age"])
    },
))
```

## Actions

Bundle mutations into reusable functions:

```go
s := state.NewStore("counter")
s.Set("count", 0)

increment := state.Action(func(ctx state.Context) error {
    current, _ := s.Get("count").(int)
    s.Set("count", current+1)
    return nil
})

// Run immediately
_ = state.Dispatch(context.Background(), increment)

// Bind to event handler
handler := state.UseAction(context.Background(), increment)
dom.RegisterHandlerFunc("increment", func() { _ = handler() })
```

## Watchers

React to store changes with watchers:

```go
s.Watch("age", func(v any) {
    log.Println("age updated", v)
})
```

Options like `WatcherDeep` and `WatcherImmediate` allow deep or immediate execution.

## Undo / Redo

Enable history with `WithHistory`:

```go
s := state.NewStore("profile", state.WithHistory(5))
s.Set("age", 30)
s.Set("age", 31)
s.Undo() // -> 30
s.Redo() // -> 31
```

## Suspense

Use Suspense to show fallbacks during async loading:

```go
var todo Todo
content := core.NewSuspense(func() (string, error) {
    if err := http.FetchJSON("/api/todo/1", &todo); err != nil {
        return "", err
    }
    return fmt.Sprintf("<div>%s</div>", todo.Title), nil
}, "<div>Loading...</div>")
```

## Binding in Components

Referencing `user/profile.fullName` in a template subscribes the component to that key. Updates trigger automatic DOM patches.

## JavaScript Interop

Expose stores to external scripts with `ExposeUpdateStore()`. Most apps can remain pure Go, but this is available for integrations.

## DevTools and Persistence

* Enable `WithDevTools` to log mutations during development.
* Enable `WithPersistence` to survive browser reloads.

---

State management in rfw combines **stores** for shared, persistent data and **signals** for local state. This approach removes the need for external libraries like Pinia—everything is built in.

## Related

* [Store Vs Signals](/docs/store-vs-signals)
