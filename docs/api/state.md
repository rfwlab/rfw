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
