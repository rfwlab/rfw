# game

```go
import "github.com/rfwlab/rfw/v2/game"
```

Game engine utilities for rfw. Provides game loop, rendering, and scene management.

## Loop

```go
import "github.com/rfwlab/rfw/v2/game/loop"
```

Game loop with fixed timestep and interpolation.

| Function/Type | Description |
| --- | --- |
| `Loop` | Main game loop struct |
| `New(fps int) *Loop` | Create a loop at given FPS |
| `Run(fn func(dt float64))` | Start loop with update function |
| `Stop()` | Stop the loop |

## Draw

```go
import "github.com/rfwlab/rfw/v2/game/draw"
```

Canvas rendering utilities.

| Function/Type | Description |
| --- | --- |
| `Canvas` | HTML5 Canvas wrapper |
| `NewCanvas(element string) *Canvas` | Create canvas from element ID |
| `Clear(color string)` | Clear canvas with color |
| `FillRect(x, y, w, h float64, color string)` | Fill rectangle |
| `DrawImage(img js.Value, x, y float64)` | Draw image |

## Scene

```go
import "github.com/rfwlab/rfw/v2/game/scene"
```

Scene graph management.

| Function/Type | Description |
| --- | --- |
| `Node` | Scene graph node |
| `NewNode() *Node` | Create a node |
| `AddChild(*Node)` | Add child node |
| `RemoveChild(*Node)` | Remove child node |
| `SetTransform(pos, rot, scale m.Vec3)` | Set transform |
| `Update(dt float64)` | Update node and children |
| `Render(*draw.Canvas)` | Render node and children |