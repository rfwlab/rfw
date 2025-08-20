# animation

Helpers for driving simple animations without the Web Animations API.

- `Translate(sel, from, to, dur)` moves elements.
- `Fade(sel, from, to, dur)` adjusts opacity.
- `Scale(sel, from, to, dur)` scales elements.
- `ColorCycle(sel, colors, dur)` cycles background colors.

## Usage

Call the animation helpers with a CSS selector, starting and ending values,
and a duration. They can be invoked from event handlers registered in the DOM.

## Example

```go
dom.RegisterHandlerFunc("animateFade", func() {
        anim.Fade("#fadeBox", 1, 0, 500*time.Millisecond)
})
```

1. `dom.RegisterHandlerFunc` links the `animateFade` ID to a Go function.
2. When triggered, `anim.Fade` fades `#fadeBox` from opaque to transparent
   over half a second.
