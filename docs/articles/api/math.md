# math

Vector and matrix helpers for graphics transformations.

| Type | Description |
| --- | --- |
| `Vec2` | 2D vector with `Add` and `Normalize`. |
| `Vec3` | 3D vector with `Add` and `Normalize`. |
| `Mat4` | 4x4 matrix supporting multiplication and projection builders. |

## Model-View-Projection

Calculating a model-view-projection matrix and uploading it to WebGL:

```go
import (
    stdmath "math"
    m "github.com/rfwlab/rfw/v1/math"
    js "github.com/rfwlab/rfw/v1/js"
    webgl "github.com/rfwlab/rfw/v1/webgl"
)

proj := m.Perspective(float32(stdmath.Pi/4), width/height, 0.1, 100)
view := m.Translation(m.Vec3{0, 0, -5})
model := m.Scale(m.Vec3{1, 1, 1})
mvp := proj.Mul(view).Mul(model)

loc := ctx.GetUniformLocation(prog, "u_mvp")
arr := js.Float32Array().New(len(mvp))
for i, v := range mvp {
    arr.SetIndex(i, v)
}
ctx.Call("uniformMatrix4fv", loc, false, arr)
```

### Prerequisites

Use when performing 2D or 3D transformations in WebGL components.

### How

1. Build projection with `math.Perspective` or `math.Orthographic`.
2. Create model and view matrices using `math.Translation` and `math.Scale`.
3. Combine them via `Mat4.Mul`.
4. Convert the matrix to a typed array with `js.Float32Array().New(len(mvp))` and upload it via `Context.Call("uniformMatrix4fv", ...)`.

### Notes and Limitations

- Matrices are column-major and use `float32` precision.
- Only basic operations are provided.
- Arrays created with `js.Float32Array().New` are managed by JavaScript and need no manual release.

### Related links

- [webgl](webgl)
