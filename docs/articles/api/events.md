# events

Utilities for handling browser events and observing DOM changes.

| Function | Description |
| --- | --- |
| `On(event, target, handler, opts...)` | Attaches a listener and returns a cleanup function. |
| `OnClick(target, handler)` | Convenience wrapper for `click` events. |
| `OnScroll(target, handler)` | Convenience wrapper for `scroll` events. |
| `OnInput(target, handler)` | Convenience wrapper for `input` events. |
| `OnTimeUpdate(target, handler)` | Convenience wrapper for `timeupdate` events. |
| `OnKeyDown(handler)` | Registers a `keydown` handler on `window`. |
| `OnKeyUp(handler)` | Registers a `keyup` handler on `window`. |
| `Listen(event, target)` | Returns a channel that receives the event's first argument. |
| `ObserveMutations(selector)` | Watches DOM mutations, skipping elements marked with `data-rfw-ignore`. |
| `ObserveIntersections(selector, opts)` | Streams `IntersectionObserverEntry` values. |

`On` registers callbacks directly and returns a function to remove the
listener when no longer needed. `Listen` converts native DOM events into
Go channels. For application state changes use reactive stores which
emit their own notifications – a separate mechanism from DOM events.

## Usage

`On` attaches a handler and provides a cleanup function:

```go
doc := dom.Doc()
stop := events.OnClick(doc.ByID("btn").Value, func(evt js.Value) {
        js.Console().Call("log", "clicked")
})
// ...
stop()
```

`Listen` turns browser events into Go channels when you prefer to range
over events concurrently.

This example listens for browser events and reacts in Go.

@include:ExampleFrame:{code:"/examples/components/event_listener_component.go", uri:"/examples/event/listener"}
