# State Management

Reactivity in rfw is driven by **stores**. A store groups values by
module and name. Components subscribe to specific keys and are
automatically re‑rendered when those values change. This section covers
the primitives available to manage state.

## Creating a store

```go
import "github.com/rfwlab/rfw/v1/state"

s := state.NewStore("profile", state.WithModule("user"))
s.Set("first", "Ada")
s.Set("last", "Lovelace")
```

Stores can be namespaced with `WithModule` and configured to persist to
`localStorage` using `WithPersistence`. The global store manager keeps
track of every registered store.

## Computed values

Computed values derive new data from existing keys. They are lazily
re‑evaluated when any dependency changes:

```go
state.Map2(s, "fullName", "first", "last", func(first, last string) string {
    return first + " " + last
})
```

Components can bind to `fullName` like any other key and the value stays
up to date as `first` or `last` changes.

For more control, `RegisterComputed` lets you derive values from any number
of dependencies. It receives a map of the watched keys and can return any
type:

```go
import "fmt"

if s.Get("profile") == nil {
    s.RegisterComputed(state.NewComputed(
        "profile",
        []string{"first", "last", "age"},
        func(m map[string]any) any {
            first, _ := m["first"].(string)
            last, _ := m["last"].(string)
            age, _ := m["age"].(int)
            return fmt.Sprintf("%s %s (%d)", first, last, age)
        },
    ))
}
```

This approach is useful for advanced cases such as combining many fields or
transforming data before exposing it to components.

## Watchers

Watchers trigger callbacks whenever a key (or set of keys) changes. Use
them for side effects such as logging or synchronising with external
systems:

```go
s.Watch("age", func(v any) {
    log.Println("age updated", v)
})
```

Options like `WatcherDeep` or `WatcherImmediate` enable more advanced
behaviour such as deep watching of nested structures or running the
callback immediately after registration.

## Binding in components

When a component references `user/profile.fullName` in its template the
framework subscribes it to that key. Any update to the store triggers a
DOM patch that keeps the rendered output in sync.

## Interacting from JavaScript

Call `ExposeUpdateStore()` to allow external scripts to mutate stores
from JavaScript. Most applications can stay entirely in Go, but this hook
is available for interoperability with existing libraries.

## Debugging and persistence

Enable `WithDevTools` when creating a store to log every mutation during
development. Persistence can be toggled with `WithPersistence` to store
state in the browser between sessions.

See the example component for a practical demonstration of computed values and reactive updates.
Stores provide reactive state management.

@include:ExampleFrame:{code:"/examples/components/state_management_component.go", uri:"/examples/state"}
