# Animations

## Why
Animations provide visual feedback and help draw attention to state changes. The [Animation API](../api/animation) runs frame-by-frame updates with helpers like `Translate`.

```go
anim.Translate("#box", anim.Point{X:0, Y:0}, anim.Point{X:100, Y:0}, time.Second)
```

## When to Use
Use these helpers when you need scripted motion beyond what CSS transitions provide.

```go
anim.Fade("#notice", time.Second)
```

## When Not to Use
Avoid the animation package for simple hover effects that CSS can handle more efficiently.

```css
.button:hover { transform: scale(1.1); }
```

## Interactive Demo
@include:ExampleFrame:{code:"/examples/components/animation_component.go", uri:"/examples/animations"}
