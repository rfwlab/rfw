# math

```go
import "github.com/rfwlab/rfw/v2/math"
```

Vector and matrix math utilities.

## Vec2

```go
type Vec2 struct{ X, Y float32 }
```

| Method | Description |
| --- | --- |
| `Add(Vec2) Vec2` | Component-wise add |
| `Sub(Vec2) Vec2` | Component-wise subtract |
| `Mul(f float32) Vec2` | Scale by factor |
| `Dot(Vec2) float32` | Dot product |
| `Length() float32` | Vector length |
| `Normalize() Vec2` | Unit vector |
| `Distance(Vec2) float32` | Distance to other |

## Vec3

```go
type Vec3 struct{ X, Y, Z float32 }
```

| Method | Description |
| --- | --- |
| `Add(Vec3) Vec3` | Component-wise add |
| `Sub(Vec3) Vec3` | Component-wise subtract |
| `Mul(f float32) Vec3` | Scale by factor |
| `Dot(Vec3) float32` | Dot product |
| `Cross(Vec3) Vec3` | Cross product |
| `Length() float32` | Vector length |
| `Normalize() Vec3` | Unit vector |

## Mat4

```go
type Mat4 [16]float32  // Column-major 4x4 matrix
```

| Function | Description |
| --- | --- |
| `Identity() Mat4` | Identity matrix |
| `Translation(Vec3) Mat4` | Translation matrix |
| `Scale(Vec3) Mat4` | Scale matrix |
| `RotationX/Y/Z(float32) Mat4` | Rotation matrices |
| `Perspective(...) Mat4` | Perspective projection |
| `LookAt(...) Mat4` | View matrix |

## Quat

```go
type Quat struct{ X, Y, Z, W float32 }
```

Quaternion for 3D rotation.

| Method | Description |
| --- | --- |
| `Identity() Quat` | Identity quaternion |
| `FromAxisAngle(Vec3, float32) Quat` | From axis-angle |
| `Mul(Quat) Quat` | Multiply quaternions |
| `Slerp(Quat, float32) Quat` | Spherical interpolation |
| `ToMat4() Mat4` | Convert to matrix |

## Additional

| Function | Description |
| --- | --- |
| `Clamp(v, min, max float32) float32` | Clamp value |
| `Lerp(a, b, t float32) float32` | Linear interpolation |
| `Rand() float32` | Random 0-1 |
| `Randn() float32` | Random normal distribution |