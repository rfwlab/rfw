// v1/ai/pathfinding/pathfinding.go
package pathfinding

import (
	"container/heap"
	"context"
	"errors"
	"math"
	"sync"

	m "github.com/rfwlab/rfw/v1/math"
)

// Grid represents a 2D grid where 0 cells are walkable and 1 are blocked.
type Grid [][]int

// NavMesh represents a polygon adjacency graph for navigation meshes.
type NavMesh struct {
	Polygons map[int]Poly
}

// Poly is a polygon node in a NavMesh.
type Poly struct {
	ID        int
	Center    m.Vec2
	Neighbors []int
}

// -------- Results (explicit outcomes: found vs canceled vs invalid) --------

type GridPathResult struct {
	Path  []m.Vec2
	Found bool
	Err   error // Err == context.Canceled on cancel; other validation errors possible
}

type MeshPathResult struct {
	Path  []int
	Found bool
	Err   error // Err == context.Canceled on cancel; other validation errors possible
}

// -------- Asynchronous pathfinder with cancelable requests --------

type Pathfinder struct {
	mu      sync.Mutex
	nextID  int
	cancels map[int]context.CancelFunc
}

// New returns a new Pathfinder.
func New() *Pathfinder {
	return &Pathfinder{cancels: make(map[int]context.CancelFunc)}
}

// RequestGridPath starts an asynchronous A* search on a grid.
// It returns a request ID and a channel that will receive the result.
func (p *Pathfinder) RequestGridPath(ctx context.Context, grid Grid, start, goal m.Vec2) (int, <-chan GridPathResult) {
	id, ctx := p.prepare(ctx)
	ch := make(chan GridPathResult, 1)
	go func() {
		defer close(ch)
		path, found, err := gridAStar(ctx, grid, start, goal)
		// If cancellation happened right before sending, surface it explicitly.
		if ctx.Err() != nil && err == nil {
			err = context.Canceled
			found = false
			path = nil
		}
		ch <- GridPathResult{Path: path, Found: found, Err: err}
		p.finish(id)
	}()
	return id, ch
}

// RequestNavMeshPath starts an asynchronous A* search on a navmesh.
// It returns a request ID and a channel that will receive the result.
func (p *Pathfinder) RequestNavMeshPath(ctx context.Context, mesh NavMesh, start, goal int) (int, <-chan MeshPathResult) {
	id, ctx := p.prepare(ctx)
	out := make(chan MeshPathResult, 1)
	go func() {
		defer close(out)
		path, found, err := navMeshAStar(ctx, mesh, start, goal)
		if ctx.Err() != nil && err == nil {
			err = context.Canceled
			found = false
			path = nil
		}
		out <- MeshPathResult{Path: path, Found: found, Err: err}
		p.finish(id)
	}()
	return id, out
}

// Cancel aborts a pending request.
func (p *Pathfinder) Cancel(id int) {
	p.mu.Lock()
	if c, ok := p.cancels[id]; ok {
		c()
		delete(p.cancels, id)
	}
	p.mu.Unlock()
}

// prepare registers a cancelable context for a new request.
func (p *Pathfinder) prepare(ctx context.Context) (int, context.Context) {
	p.mu.Lock()
	p.nextID++
	id := p.nextID
	ctx, cancel := context.WithCancel(ctx)
	p.cancels[id] = cancel
	p.mu.Unlock()
	return id, ctx
}

// finish cleans up tracking for a completed/canceled request.
func (p *Pathfinder) finish(id int) { p.Cancel(id) }

// -------- Grid A* --------

var (
	// ErrOutOfBounds indicates start/goal outside the grid or empty grid.
	ErrOutOfBounds = errors.New("start/goal out of bounds")
	// ErrBlocked indicates start/goal is on a non-walkable cell.
	ErrBlocked = errors.New("start/goal is blocked")
)

type gridNode struct {
	pos    m.Vec2
	g, f   float32
	parent *gridNode
	index  int
}

type priorityQueue []*gridNode

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index, pq[j].index = i, j }
func (pq *priorityQueue) Push(x any)        { *pq = append(*pq, x.(*gridNode)) }
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[:n-1]
	return x
}

// inBounds checks if (x,y) is inside the grid.
func inBounds(grid Grid, x, y int) bool {
	return y >= 0 && y < len(grid) && x >= 0 && x < len(grid[0])
}

// gridAStar performs A* on a 4-neighborhood grid with unit costs and Manhattan heuristic.
func gridAStar(ctx context.Context, grid Grid, start, goal m.Vec2) ([]m.Vec2, bool, error) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return nil, false, ErrOutOfBounds
	}
	sx, sy := int(start.X), int(start.Y)
	gx, gy := int(goal.X), int(goal.Y)
	if !inBounds(grid, sx, sy) || !inBounds(grid, gx, gy) {
		return nil, false, ErrOutOfBounds
	}
	if grid[sy][sx] == 1 || grid[gy][gx] == 1 {
		return nil, false, ErrBlocked
	}
	if sx == gx && sy == gy {
		return []m.Vec2{{float32(sx), float32(sy)}}, true, nil
	}

	// Manhattan heuristic (admissible in 4-dir grid with unit costs).
	h := func(x, y int) float32 { return float32(math.Abs(float64(x-gx)) + math.Abs(float64(y-gy))) }

	startNode := &gridNode{pos: m.Vec2{float32(sx), float32(sy)}, g: 0, f: h(sx, sy)}
	open := priorityQueue{startNode}
	heap.Init(&open)
	visited := map[[2]int]*gridNode{{sx, sy}: startNode}

	for open.Len() > 0 {
		select {
		case <-ctx.Done():
			return nil, false, context.Canceled
		default:
		}
		current := heap.Pop(&open).(*gridNode)
		x, y := int(current.pos.X), int(current.pos.Y)
		if x == gx && y == gy {
			return reconstructGrid(current), true, nil
		}
		// 4-neighbors
		for _, n := range [][2]int{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1}} {
			nx, ny := n[0], n[1]
			if !inBounds(grid, nx, ny) || grid[ny][nx] == 1 {
				continue
			}
			gScore := current.g + 1
			key := [2]int{nx, ny}
			if v, ok := visited[key]; ok && gScore >= v.g {
				continue
			}
			node := &gridNode{
				pos:    m.Vec2{float32(nx), float32(ny)},
				g:      gScore,
				f:      gScore + h(nx, ny),
				parent: current,
			}
			visited[key] = node
			heap.Push(&open, node)
		}
	}
	// Exhausted without reaching goal: not found.
	return nil, false, nil
}

// reconstructGrid rebuilds the path from the goal node by following parents.
func reconstructGrid(n *gridNode) []m.Vec2 {
	var path []m.Vec2
	for n != nil {
		path = append(path, n.pos)
		n = n.parent
	}
	// reverse in-place
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

// -------- NavMesh A* (uniform neighbor cost + Euclidean heuristic on polygon centers) --------

type navNode struct {
	id     int
	g, f   float32
	parent *navNode
	index  int
}

type priorityQueueNav []*navNode

func (pq priorityQueueNav) Len() int           { return len(pq) }
func (pq priorityQueueNav) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq priorityQueueNav) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}
func (pq *priorityQueueNav) Push(x any) { *pq = append(*pq, x.(*navNode)) }
func (pq *priorityQueueNav) Pop() any {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[:n-1]
	return x
}

// navMeshAStar performs A* on a polygon adjacency graph.
// Cost is uniform per edge; heuristic is Euclidean distance between polygon centers.
func navMeshAStar(ctx context.Context, mesh NavMesh, start, goal int) ([]int, bool, error) {
	if mesh.Polygons == nil {
		return nil, false, errors.New("empty navmesh")
	}
	if _, ok := mesh.Polygons[start]; !ok {
		return nil, false, errors.New("start polygon not found")
	}
	if _, ok := mesh.Polygons[goal]; !ok {
		return nil, false, errors.New("goal polygon not found")
	}
	if start == goal {
		return []int{start}, true, nil
	}

	// Euclidean heuristic on polygon centers.
	h := func(a int) float32 {
		sa := mesh.Polygons[a].Center
		sb := mesh.Polygons[goal].Center
		dx := sa.X - sb.X
		dy := sa.Y - sb.Y
		return float32(math.Hypot(float64(dx), float64(dy)))
	}

	startNode := &navNode{id: start, g: 0, f: h(start)}
	open := priorityQueueNav{startNode}
	heap.Init(&open)
	visited := map[int]*navNode{start: startNode}

	for open.Len() > 0 {
		select {
		case <-ctx.Done():
			return nil, false, context.Canceled
		default:
		}
		current := heap.Pop(&open).(*navNode)
		if current.id == goal {
			return reconstructNav(current), true, nil
		}
		for _, nb := range mesh.Polygons[current.id].Neighbors {
			// Defensive: skip neighbors not present in the mesh.
			if _, ok := mesh.Polygons[nb]; !ok {
				continue
			}
			gScore := current.g + 1 // uniform edge cost
			if v, ok := visited[nb]; ok && gScore >= v.g {
				continue
			}
			node := &navNode{id: nb, g: gScore, f: gScore + h(nb), parent: current}
			visited[nb] = node
			heap.Push(&open, node)
		}
	}
	// Exhausted without reaching goal: not found.
	return nil, false, nil
}

// reconstructNav rebuilds the polygon path by following parents.
func reconstructNav(n *navNode) []int {
	var path []int
	for n != nil {
		path = append(path, n.id)
		n = n.parent
	}
	// reverse in-place
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}
