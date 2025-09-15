# Pathfinding

The **pathfinding** package provides A\* search on grids and navmeshes, helping agents navigate 2D worlds.

## Why Use It

Deterministic pathfinding enables predictable and repeatable movement for characters and AI. Use it when moving entities on tile-based maps or polygonal navmeshes.

## Getting Started

1. Create a pathfinder with `pathfinding.New()`
2. Call `RequestGridPath` or `RequestNavMeshPath` to start a search
3. Read the resulting path from the returned channel, or cancel with `Cancel`

```go
pf := pathfinding.New()
id, ch := pf.RequestGridPath(ctx, grid, start, goal)
// ...
pf.Cancel(id)
path := <-ch
```

### Sample Grid

```
0 0 0
1 1 0
0 0 0
```

Cells marked `1` are obstacles; `0` are walkable.

## Example: NavMesh Path

Request a path across polygons:

```go
mesh := pathfinding.NavMesh{Polygons: map[int]pathfinding.Poly{
    1: {ID: 1, Center: math.Vec2{0, 0}, Neighbors: []int{2}},
    2: {ID: 2, Center: math.Vec2{1, 0}, Neighbors: []int{1, 3}},
    3: {ID: 3, Center: math.Vec2{2, 0}, Neighbors: []int{2}},
}}
pf := pathfinding.New()
_, ch := pf.RequestNavMeshPath(ctx, mesh, 1, 3)
path := <-ch
```

@include\:ExampleFrame:{code:"/examples/components/pathfinding\_component.go", uri:"/examples/pathfinding"}

## Notes

* Grid cells with value `1` are treated as obstacles
* NavMesh searches return polygon IDs, not exact edge crossings

## Related

* [math](../api/math)
* [testing](../testing)
