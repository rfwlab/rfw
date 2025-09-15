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

