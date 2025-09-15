# scene

Helpers for scene graph traversal and updates.

| Function | Description |
| --- | --- |
| `NewNode() *Node` | Create a scene node. |
| `Traverse(n *Node, fn func(*Node))` | Walk the scene graph. |
| `Update(root *Node, t Ticker)` | Update nodes each frame using a ticker. |

