# Scene Graph

The **scene graph** organizes objects in hierarchies so updates and rendering cascade through parents and children.

---

## When to Use

Use a scene graph when you need:

* Hierarchical transforms (e.g. moving a parent moves its children)
* Efficient updates for many entities per frame

For simple single-object scenes, you can skip it.

---

## How It Works

1. Create nodes with optional entities and components
2. Connect nodes with `AddChild` to form a hierarchy
3. Call `scene.Update` every frame, usually inside the [game loop](/articles/api/game-loop)

---

## API Overview

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

---

## Example

The following graph separates units and buildings, then updates them every frame:

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

---

## Notes

* Traversal is depth-first
* No built-in removal helpers—manage node lifecycles manually

---

## Related

* [Game loop](/articles/api/game-loop)
