# state

Centralized reactive data stores.

| Function | Description |
| --- | --- |
| `NewStore(name, opts...)` | Create a store. |
| `Set(key, value)` | Update a key and trigger bindings. |
| `RegisterComputed(comp)` | Define derived values. |
| `RegisterWatcher(w)` | Run a callback after changes. |
| `ExposeUpdateStore()` | Expose `goUpdateStore` to JavaScript. |

Stores are the primary mechanism for application state. They emit
notifications to any component that reads their values, keeping most apps
free of handwritten JavaScript.

## Usage

Stores are created with `state.NewStore` and read or updated via `Get` and
`Set`. Derived values can be defined with `RegisterComputed`.

## Example

```go
if store.Get("double") == nil {
        store.RegisterComputed(state.NewComputed("double", []string{"count"}, func(m map[string]any) any {
                if v, ok := m["count"].(int); ok {
                        return v * 2
                }
                return 0
        }))
}

func (c *StateManagementComponent) Increment() {
        if v, ok := c.Store.Get("count").(int); ok {
                c.Store.Set("count", v+1)
        }
}
```

1. `RegisterComputed` defines the derived value `double` based on `count`.
2. The `Increment` function reads `count` and updates it with `Set`.
3. Each change automatically updates parts of the UI that depend on these
   values.
