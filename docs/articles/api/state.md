# state

```go
import "github.com/rfwlab/rfw/v2/state"
```

Centralized reactive data stores.

## Store Creation

| Function | Description |
| --- | --- |
| `NewStore(name string, opts ...StoreOption) *Store` | Create a new store. |
| `WithModule(m string) StoreOption` | Set the module namespace. |
| `WithHistory(limit int) StoreOption` | Enable undo/redo up to `limit` steps. |
| `WithPersistence(path string) StoreOption` | Persist store state to disk. |

## Store Methods

| Method | Description |
| --- | --- |
| `Get(key string) any` | Retrieve a key's value. |
| `Set(key string, value any)` | Update a key and notify listeners. |
| `OnChange(key string, fn func(any)) func()` | Listen for changes; returns cleanup. |
| `Undo()` | Revert last mutation (history required). |
| `Redo()` | Reapply last undone mutation (history required). |
| `Module() string` | Current module namespace. |
| `Name() string` | Store identifier. |
| `Snapshot() map[string]any` | Copy of current key/value state. |

## GlobalStoreManager

| Method | Description |
| --- | --- |
| `GlobalStoreManager.Snapshot() map[string]map[string]any` | Deep copy of all stores and state. |
| `GlobalStoreManager.UnregisterStore(module, name)` | Remove a store from the manager. |

## Signals

| Function | Description |
| --- | --- |
| `NewSignal[T any](initial T) *Signal[T]` | Create a fine-grained reactive value. |
| `(Signal[T]).Get() T` | Read current value; tracks in effects. |
| `(Signal[T]).Set(v T)` | Update value; re-run dependents. |

## Helpers

| Function | Description |
| --- | --- |
| `Watch(fn func()) func()` | Re-run `fn` when signals it reads change; returns stop. |
| `Path(parts ...string) string` | Build a dot-separated store key path. |

## Type Aliases

| Type | Definition |
| --- | --- |
| `Int` | `Signal[int]` |
| `String` | `Signal[string]` |
| `Bool` | `Signal[bool]` |
| `Float` | `Signal[float64]` |
| `Any` | `Signal[any]` |