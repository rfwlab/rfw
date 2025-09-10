# Scene Graph

## Why

Games often organise objects in hierarchies. A scene graph groups nodes so updates and rendering can cascade through parents and children.

## When to Use

Use a scene graph when you need hierarchical transforms or want to update many entities each frame. Simple single-object scenes can skip it.

## How

1. Create nodes with optional entities and components.
2. Connect nodes via `AddChild` to form the hierarchy.
3. Call `scene.Update` every frame, typically from the [game loop](/articles/api/game-loop).

## API

```go
package scene

func NewNode() *Node
func (n *Node) AddChild(child *Node)
func (n *Node) AddEntity(e Entity)
func Traverse(n *Node, fn func(*Node))
func Update(root *Node, t Ticker)

type Transform struct { X, Y float64 }
type Component interface { Update(*Node, Ticker) }
type Entity interface { Components() []Component }
type Ticker struct { Delta time.Duration; FPS float64 }
```

## Example

The following scene graph splits units and buildings while updating them every frame:

```go
root := scene.NewNode()
units := scene.NewNode()
buildings := scene.NewNode()
root.AddChild(units)
root.AddChild(buildings)

soldier := scene.NewNode()
barracks := scene.NewNode()
units.AddChild(soldier)
buildings.AddChild(barracks)

move := &MoveComponent{}
unit := &UnitEntity{comps: []scene.Component{move}}
soldier.AddEntity(unit)

 loop.OnUpdate(func(t loop.Ticker) {
     scene.Update(root, scene.Ticker{Delta: t.Delta, FPS: t.FPS})
 })
loop.Start()
```

## Notes and Limitations

Traversal is depth-first and no removal helpers are provided. Manage memory and node lifecycles in your own code.

## Related Links

- [game loop](/articles/api/game-loop)
