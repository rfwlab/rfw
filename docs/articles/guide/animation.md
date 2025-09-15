# Animations

Animations make interfaces feel alive. They provide feedback, guide attention, and can improve the overall user experience. In **rfw**, animations are driven from Go code through the [Animation API](../api/animation). This API works frame‑by‑frame and exposes helpers such as `Translate`, `Fade`, and keyframe utilities.

## Using Translate

The `Translate` helper moves an element from one position to another over a specified duration.

```go
import (
    "time"
    anim "github.com/rfwlab/rfw/v1/animation"
)

anim.Translate("#box", anim.Point{X:0, Y:0}, anim.Point{X:100, Y:0}, time.Second)
```

This code animates the element with ID `box` from `(0,0)` to `(100,0)` in one second.

## Using Fade

The `Fade` helper smoothly hides an element by animating its opacity:

```go
anim.Fade("#notice", time.Second)
```

This fades out the element with ID `notice` in one second.

## When to Use rfw Animations

Use the animation package when you need motion that depends on your application state, or when CSS alone isn’t enough. For example:

* Animate an element after data is loaded.
* Run a sequence of movements in response to a signal.
* Coordinate multiple UI parts programmatically.

## When to Use CSS Instead

Keep simple interactions in CSS—it’s faster and more lightweight:

```css
.button:hover {
  transform: scale(1.1);
}
```

Use CSS for hover, focus, or basic transitions that don’t need runtime logic.

## Interactive Example

The following demo shows how animations integrate with components:

@include\:ExampleFrame:{code:"/examples/components/animation\_component.go", uri:"/examples/animations"}

---

Animations are a powerful tool, but like any feature they should be applied with care: use them to highlight state changes and improve clarity, not to distract.
