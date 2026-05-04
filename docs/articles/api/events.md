# events

```go
import "github.com/rfwlab/rfw/v2/events"
```

Browser event handling and DOM observation.

## On-family

| Function | Description |
| --- | --- |
| `On(event string, target *Element, handler func(*Event), opts ...EventOption) func()` | Attach listener; returns cleanup. |
| `OnClick(target *Element, handler func(*Event)) func()` | Click event shorthand. |
| `OnInput(target *Element, handler func(*Event)) func()` | Input event shorthand. |
| `OnKeyDown(handler func(*Event)) func()` | Keydown on window. |
| `OnKeyUp(handler func(*Event)) func()` | Keyup on window. |
| `OnScroll(target *Element, handler func(*Event)) func()` | Scroll event shorthand. |
| `OnTimeUpdate(target *Element, handler func(*Event)) func()` | Timeupdate event shorthand. |

## Observation

| Function | Description |
| --- | --- |
| `Listen(event string, target *Element) <-chan any` | Channel that receives event payloads. |
| `ObserveMutations(selector string) <-chan MutationRecord` | Watch DOM mutations (skips `data-rfw-ignore`). |
| `ObserveIntersections(selector string, opts IntersectionObserverInit) <-chan IntersectionObserverEntry` | Stream intersection entries. |