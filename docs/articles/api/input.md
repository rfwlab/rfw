# input

Keyboard and mouse mapping with drag tracking and camera helpers.

| Function | Description |
| --- | --- |
| `New()` | Creates a manager and wires browser events. |
| `(*Manager).BindKey(action, key)` | Map a keyboard key to an action. |
| `(*Manager).RebindKey(action, key)` | Change the key for an action. |
| `(*Manager).BindMouse(action, button)` | Map a mouse button to an action. |
| `(*Manager).RebindMouse(action, button)` | Change the button for an action. |
| `(*Manager).IsActive(action)` | Query whether an action is engaged. |
| `(*Manager).DragRect()` | Current drag start and end positions. |
| `(*Manager).Camera()` | Snapshot of the camera state. |
| `(*Manager).Pan(dx, dy)` | Translate the camera. |
| `(*Manager).Zoom(delta)` | Adjust camera zoom. |
| `(*Manager).Rotate(delta)` | Adjust camera rotation. |

