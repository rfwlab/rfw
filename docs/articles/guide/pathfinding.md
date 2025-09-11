# Pathfinding

Grid and navmesh A\* searches help agents navigate 2D worlds.

## Why

Deterministic pathfinding enables predictable movement for characters and AI.

## Prerequisites

Use pathfinding when moving entities in tile or polygon based maps.

## How

1. Create a `Pathfinder` with `pathfinding.New()`.
2. Call `RequestGridPath` or `RequestNavMeshPath` to start a search.
3. Read the resulting path from the returned channel or cancel it with `Cancel`.

```go
pf := pathfinding.New()
id, ch := pf.RequestGridPath(ctx, grid, start, goal)
// ...
pf.Cancel(id)
path := <-ch
```

### Sample grid

```
0 0 0
1 1 0
0 0 0
```

### APIs used

- `pathfinding.New`
- `(*Pathfinder).RequestGridPath`
- `(*Pathfinder).RequestNavMeshPath`
- `(*Pathfinder).Cancel`
- `math.Vec2`

## Example

Request a navmesh path between polygons:

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

Explore the grid search demo below.

@include:ExampleFrame:{code:"/examples/components/pathfinding_component.go", uri:"/examples/pathfinding"}

## Notes and Limitations

- Grid cells with value `1` are treated as obstacles.
- Navmesh paths return polygon IDs rather than exact edge crossings.

## Related links

- [math](../api/math)
- [testing](../testing)
