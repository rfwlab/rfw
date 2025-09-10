package scene

import "time"

// Transform represents a 2D position for a node.
type Transform struct {
	X, Y float64
}

// Component defines behavior that can update each frame.
type Component interface {
	Update(*Node, Ticker)
}

// Entity groups a set of components.
type Entity interface {
	Components() []Component
}

// Node is a scene graph node with a transform and children.
type Node struct {
	Transform Transform
	Children  []*Node
	Entities  []Entity
}

// NewNode creates an empty node.
func NewNode() *Node {
	return &Node{}
}

// AddChild appends a child node.
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

// AddEntity attaches an entity to the node.
func (n *Node) AddEntity(e Entity) {
	n.Entities = append(n.Entities, e)
}

// Traverse walks the scene graph depth-first, invoking fn for each node.
func Traverse(n *Node, fn func(*Node)) {
	if n == nil {
		return
	}
	fn(n)
	for _, c := range n.Children {
		Traverse(c, fn)
	}
}

// Update traverses the graph and updates all components on every node.
func Update(root *Node, t Ticker) {
	Traverse(root, func(n *Node) {
		for _, e := range n.Entities {
			for _, c := range e.Components() {
				c.Update(n, t)
			}
		}
	})
}

// Ticker mirrors the game loop ticker with delta time and FPS.
type Ticker struct {
	Delta time.Duration
	FPS   float64
}
