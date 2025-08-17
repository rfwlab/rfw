# Animations

`v1/animation` offers helpers that drive `requestAnimationFrame` loops:

```go
anim.Translate("#box", anim.Point{X:0, Y:0}, anim.Point{X:100, Y:0}, time.Second)
```

Helpers like `Fade`, `Scale` and `ColorCycle` operate on elements selected by CSS selectors.
