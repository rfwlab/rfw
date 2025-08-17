# State Management

Stores group reactive values by module and name. Components can subscribe
to store values and automatically re-render when data changes.

```go
s := state.NewStore("profile", state.WithModule("user"))
s.RegisterComputed(state.NewComputed("fullName", []string{"first","last"}, func(m map[string]interface{}) interface{} {
    return m["first"].(string)+" "+m["last"].(string)
}))
```

Watchers trigger callbacks whenever a value changes:

```go
s.Watch("age", func(v interface{}) {
    log.Println("age updated", v)
})
```

Call `ExposeUpdateStore()` to allow external scripts to mutate stores
from JavaScript, though most applications can stay entirely in Go.
