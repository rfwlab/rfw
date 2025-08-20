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
