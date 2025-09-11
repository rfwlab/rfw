package pathfinding

import (
	"context"
	"testing"

	m "github.com/rfwlab/rfw/v1/math"
)

func TestRequestGridPathFindsPath(t *testing.T) {
	pf := New()
	grid := Grid{
		{0, 0, 0},
		{1, 1, 0},
		{0, 0, 0},
	}
	start := m.Vec2{0, 0}
	goal := m.Vec2{2, 2}
	_, ch := pf.RequestGridPath(context.Background(), grid, start, goal)
	res := <-ch
	if !res.Found {
		t.Fatalf("expected path, got err %v", res.Err)
	}
	if len(res.Path) == 0 || res.Path[0] != start || res.Path[len(res.Path)-1] != goal {
		t.Fatalf("unexpected path %v", res.Path)
	}
}

func TestRequestNavMeshPathFindsPath(t *testing.T) {
	mesh := NavMesh{Polygons: map[int]Poly{
		1: {ID: 1, Center: m.Vec2{0, 0}, Neighbors: []int{2}},
		2: {ID: 2, Center: m.Vec2{1, 0}, Neighbors: []int{1, 3}},
		3: {ID: 3, Center: m.Vec2{2, 0}, Neighbors: []int{2}},
	}}
	pf := New()
	_, ch := pf.RequestNavMeshPath(context.Background(), mesh, 1, 3)
	res := <-ch
	if !res.Found {
		t.Fatalf("expected path, got err %v", res.Err)
	}
	if len(res.Path) != 3 || res.Path[0] != 1 || res.Path[2] != 3 {
		t.Fatalf("unexpected path %v", res.Path)
	}
}

func TestCancelGridPath(t *testing.T) {
	pf := New()
	// Large grid to give time for cancellation.
	grid := make(Grid, 100)
	for i := range grid {
		grid[i] = make([]int, 100)
	}
	start := m.Vec2{0, 0}
	goal := m.Vec2{99, 99}
	id, ch := pf.RequestGridPath(context.Background(), grid, start, goal)
	pf.Cancel(id)
	res := <-ch
	if res.Err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", res.Err)
	}
}
