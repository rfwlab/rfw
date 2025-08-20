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

State stores drive UI updates in this example.

@include:ExampleFrame:{code:"/examples/components/state_management_component.go", uri:"/examples/state"}
