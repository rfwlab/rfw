# state

Centralized reactive data stores.

| Function | Description |
| --- | --- |
| `NewStore(name, opts...)` | Create a store. |
| `Set(key, value)` | Update a key and trigger bindings. |
| `RegisterComputed(comp)` | Define derived values. |
| `Map(store, key, dep, fn)` | Helper to map one key to another. |
| `Map2(store, key, depA, depB, fn)` | Map two keys into a derived value. |
| `RegisterWatcher(w)` | Run a callback after changes. |
| `ExposeUpdateStore()` | Expose `goUpdateStore` to JavaScript. |

Stores are the primary mechanism for application state. They emit
notifications to any component that reads their values, keeping most apps
free of handwritten JavaScript.

## Usage

Stores are created with `state.NewStore` and read or updated via `Get` and
`Set`. Derived values can be defined with `RegisterComputed` or convenience
helpers like `Map` and `Map2`.

## Example

```go
if store.Get("double") == nil {
        state.Map(store, "double", "count", func(v int) int {
                return v * 2
        })
}

func (c *StateManagementComponent) Increment() {
        if v, ok := c.Store.Get("count").(int); ok {
                c.Store.Set("count", v+1)
        }
}
```

1. `Map` defines the derived value `double` based on `count`.
2. The `Increment` function reads `count` and updates it with `Set`.
3. Each change automatically updates parts of the UI that depend on these
   values.

### Advanced computed values

`RegisterComputed` provides full control for deriving data from multiple
dependencies. It accepts a `Computed` definition listing the keys to watch
and a function that receives their current values in a map. Use it when the
`Map` helpers are insufficient:

```go
import "fmt"

if store.Get("profile") == nil {
        store.RegisterComputed(state.NewComputed(
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

This example combines three fields into a single formatted string, but the
function can perform any transformation and return any type.
