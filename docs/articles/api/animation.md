# animation

```go
import "github.com/rfwlab/rfw/v2/animation"
```

Animation primitives for rfw. Provides keyframe-based animation system.

## Keyframe

```go
type Keyframe struct {
    Time   float64
    Easing func(float64) float64
}
```

Animation frame at a specific time with easing function.

## Easing Functions

| Function | Description |
| --- | --- |
| `Linear(t float64) float64` | No easing |
| `EaseIn(t float64) float64` | Ease in (quadratic) |
| `EaseOut(t float64) float64` | Ease out (quadratic) |
| `EaseInOut(t float64) float64` | Ease in-out (quadratic) |
| `Bounce(t float64) float64` | Bounce effect |
| `Elastic(t float64) float64` | Elastic effect |

## Lerp Functions

| Function | Description |
| --- | --- |
| `LerpFloat(a, b, t float64) float64` | Linear interpolation |
| `LerpVec2(a, b m.Vec2, t float64) m.Vec2` | Vector interpolation |
| `LerpVec3(a, b m.Vec3, t float64) m.Vec3` | 3D vector interpolation |
| `LerpQuat(a, b m.Quat, t float64) m.Quat` | Quaternion interpolation |

## Animation Types

See [api/cinema](/api/cinema) for the cinema-style animation system.