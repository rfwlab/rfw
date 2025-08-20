# animation

Helpers for driving simple animations and for leveraging the Web Animations API.

- `Translate(sel, from, to, dur)` moves elements.
- `Fade(sel, from, to, dur)` adjusts opacity.
- `Scale(sel, from, to, dur)` scales elements.
- `ColorCycle(sel, colors, dur)` cycles background colors.
- `Keyframes(sel, frames, opts)` runs a Web Animations API sequence and
  returns the underlying `Animation` object.

## Usage

Call the animation helpers with a CSS selector, starting and ending values,
and a duration. They can be invoked from event handlers registered in the DOM.
`Keyframes` accepts arrays of frame definitions and option maps that are
passed directly to the browser's `Element.animate` API.

## Example

```go
dom.RegisterHandlerFunc("animateFade", func() {
        anim.Fade("#fadeBox", 1, 0, 500*time.Millisecond)
})

dom.RegisterHandlerFunc("animateSpin", func() {
        frames := []map[string]any{
                {"transform": "rotate(0deg)"},
                {"transform": "rotate(360deg)"},
        }
        opts := map[string]any{"duration": 500, "iterations": 1}
        anim.Keyframes("#spinBox", frames, opts)
})
```

1. `dom.RegisterHandlerFunc` links the `animateFade` ID to a Go function.
2. When triggered, `anim.Fade` fades `#fadeBox` from opaque to transparent
   over half a second.
3. `anim.Keyframes` delegates to the Web Animations API to spin `#spinBox`
   once around its center.
