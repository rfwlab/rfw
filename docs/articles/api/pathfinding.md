# pathfinding

Asynchronous A* grid and navmesh searches.

| Function | Description |
| --- | --- |
| `New() *Pathfinder` | Construct a Pathfinder. |
| `(p *Pathfinder) RequestGridPath(ctx, grid, start, goal)` | Start an asynchronous A* grid search. |
| `(p *Pathfinder) RequestNavMeshPath(ctx, mesh, start, goal)` | Start an asynchronous navmesh search. |
| `(p *Pathfinder) Cancel(id int)` | Cancel a pending request. |

