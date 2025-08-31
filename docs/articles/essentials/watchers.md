# Watchers

Watchers run callback functions in response to reactive state changes. They are ideal for side effects like logging, analytics, or interacting with external APIs when certain values mutate.

## Watching Component Fields

Inside a component, call `state.Watch` in `Init` or `OnMount` to observe a field:

```go
func (c *Counter) Init(props map[string]any) {
  c.HTMLComponent.Init(props)
  state.Watch(&c.Count, func(v any) {
    log.Println("count changed to", v)
  })
}
```

The watcher function receives the new value each time `Count` changes.

## Watching Stores

`state.Watch` also accepts store values:

```go
state.Watch(state.Path{"default", "counter", "count"}, func(v any) {
  // react to global counter
})
```

The path points to a module, store, and key. Watchers return a function to stop observing. Store this function and call it from `OnUnmount` to avoid leaks.

## Immediate and Deep Options

Advanced watchers can be configured to run immediately or to track nested structures. Pass `state.WatchOptions{Immediate: true, Deep: true}` to customize behavior.

Watchers complement computed properties by performing side effects without cluttering rendering logic.
