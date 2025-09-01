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
| `WithHistory(limit)` | Enable mutation history for `Undo`/`Redo` up to `limit` steps. |
| `Undo()` | Revert the last mutation when history is enabled. |
| `Redo()` | Reapply the last undone mutation when history is enabled. |
| `ExposeUpdateStore()` | Expose `goUpdateStore` to JavaScript. |

Stores are the primary mechanism for application state. They emit
notifications to any component that reads their values, keeping most apps
free of handwritten JavaScript.

For a comparison between store and signals, see [Stores vs signals](../guide/store-vs-signals).

## StoreHook

`state.StoreHook` runs on every mutation, allowing external observers such as
plugins or devtools to react to state changes without importing the `state`
package directly. Assign a function with the signature
`func(module, store, key string, value any)` to receive notifications for each
update.

## SetLogger

`state.SetLogger` replaces the default logger used when stores run in
development mode. Provide an implementation of `state.Logger` and call
`state.SetLogger` to capture or redirect mutation logs.

## History

Passing `WithHistory` to `state.NewStore` records mutations and enables `Undo` and `Redo`:

```go
s := state.NewStore("profile", state.WithHistory(5))
s.Set("first", "Ada")
s.Set("first", "Grace")
s.Undo() // first -> Ada
s.Redo() // first -> Grace
```

## Usage

Stores are created with `state.NewStore` and read or updated via `Get` and
`Set`. Derived values can be defined with `RegisterComputed` or convenience
helpers like `Map` and `Map2`.

State stores drive UI updates in this example.

@include:ExampleFrame:{code:"/examples/components/state_management_component.go", uri:"/examples/state"}
