package scene

import "testing"

type testComponent struct{ count int }

func (c *testComponent) Update(_ *Node, _ Ticker) { c.count++ }

type testEntity struct{ comps []Component }

func (e *testEntity) Components() []Component { return e.comps }

func TestTraverse(t *testing.T) {
	root := NewNode()
	root.AddChild(NewNode())
	root.AddChild(NewNode())

	visited := 0
	Traverse(root, func(*Node) { visited++ })
	if visited != 3 {
		t.Fatalf("expected 3 nodes visited, got %d", visited)
	}
}

func TestUpdate(t *testing.T) {
	root := NewNode()
	child := NewNode()
	root.AddChild(child)

	c1 := &testComponent{}
	c2 := &testComponent{}
	root.AddEntity(&testEntity{comps: []Component{c1}})
	child.AddEntity(&testEntity{comps: []Component{c2}})

	Update(root, Ticker{})

	if c1.count != 1 {
		t.Fatalf("expected root component updated once, got %d", c1.count)
	}
	if c2.count != 1 {
		t.Fatalf("expected child component updated once, got %d", c2.count)
	}
}
