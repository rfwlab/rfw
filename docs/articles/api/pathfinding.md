# pathfinding

Asynchronous A* grid and navmesh searches.

## Why

Deterministic navigation is essential for AI and characters moving through 2D worlds.

## When to Use

Use `pathfinding` when entities must plan routes through grids or polygon adjacency graphs.

## How

1. Construct a `Pathfinder` with `pathfinding.New()`.
2. Start a search with `RequestGridPath` or `RequestNavMeshPath`.
3. Cancel a request early with `Cancel` if the path is no longer needed.

## API

```go
package pathfinding

type Grid [][]int

type NavMesh struct {
    Polygons map[int]Poly
}

type Poly struct {
    ID        int
    Center    math.Vec2
    Neighbors []int
}

type GridPathResult struct {
    Path  []math.Vec2
    Found bool
    Err   error
}

type MeshPathResult struct {
    Path  []int
    Found bool
    Err   error
}

type Pathfinder struct{}

func New() *Pathfinder
func (p *Pathfinder) RequestGridPath(ctx context.Context, grid Grid, start, goal math.Vec2) (int, <-chan GridPathResult)
func (p *Pathfinder) RequestNavMeshPath(ctx context.Context, mesh NavMesh, start, goal int) (int, <-chan MeshPathResult)
func (p *Pathfinder) Cancel(id int)
```

## Example

```go
pf := pathfinding.New()
_, ch := pf.RequestGridPath(ctx, grid, start, goal)
path := <-ch
```

## Notes and Limitations

- `ErrOutOfBounds` and `ErrBlocked` report invalid start or goal positions.
- Cancelled searches yield `context.Canceled`.

## Related Links

- [Pathfinding guide](../guide/pathfinding)
- [math](./math)
