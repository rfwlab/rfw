# events

Utilities for observing browser events and DOM mutations.

| Function | Description |
| --- | --- |
| `Listen(event, target)` | Returns a channel that receives the event's first argument. |
| `ObserveMutations(selector)` | Watches DOM mutations. |
| `ObserveIntersections(selector, opts)` | Streams `IntersectionObserverEntry` values. |

`Listen` converts native DOM events into Go channels. For application
state changes use reactive stores which emit their own notifications – a
separate mechanism from DOM events.

## Usage

`Listen` turns browser events into Go channels. You can range over the
channel to react to events concurrently.

## Example

```go
btn := js.Document().Call("getElementById", "clickBtn")
ch := events.Listen("click", btn)
go func() {
        for range ch {
                switch v := cmp.Store.Get("clicks").(type) {
                case float64:
                        cmp.Store.Set("clicks", v+1)
                case int:
                        cmp.Store.Set("clicks", float64(v)+1)
                }
        }
}()
```

1. Obtain the button from the DOM.
2. `events.Listen` creates a channel that receives clicks on the button.
3. A goroutine increments the store whenever an event arrives.
