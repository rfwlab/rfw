# state

Centralized reactive data stores.

| Function | Description |
| --- | --- |
| `NewStore(name, opts...)` | Create a store. |
| `Get(key)` | Retrieve a key's value. |
| `Set(key, value)` | Update a key and trigger bindings. |
| `OnChange(key, listener)` | Listen for changes to a key. |
| `RegisterComputed(comp)` | Define derived values. |
| `Map(store, key, dep, fn)` | Helper to map one key to another. |
| `Map2(store, key, depA, depB, fn)` | Map two keys into a derived value. |
| `RegisterWatcher(w)` | Run a callback after changes and return a cleanup function. |
| `StoreManager.Snapshot()` | Deep copy of all stores and their current state. |
| `StoreManager.UnregisterStore(module, name)` | Remove a store from the manager. |
| `WithHistory(limit)` | Enable mutation history for `Undo`/`Redo` up to `limit` steps. |
| `Undo()` | Revert the last mutation when history is enabled. |
| `Redo()` | Reapply the last undone mutation when history is enabled. |
| `ExposeUpdateStore()` | Expose `goUpdateStore` to JavaScript for debugging or legacy scenarios. |
| `SnapshotSignals()` | Copy of all tracked signals for debugging. |
| `Action` | Function type executed with a context. |
| `Dispatch(ctx, action)` | Execute an Action with a context. |
| `UseAction(ctx, action)` | Bind an Action to a Context and return a callback. |

