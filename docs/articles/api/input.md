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

## Why

Centralizes input handling and camera state so applications can query
high‑level actions instead of raw events.

## Prerequisites

Use when reacting to keyboard and mouse input in browser builds.

## How

1. Create a manager with `input.New()`.
2. Bind actions to keys or buttons.
3. Query `IsActive` each frame or use `DragRect` and `Camera` helpers.

```go
m := input.New()
m.BindKey("shoot", "Space")
m.BindMouse("pan", 1)
if m.IsActive("shoot") {
        // fire weapon
}
start, end, dragging := m.DragRect()
_ = start
_ = end
_ = dragging
```

## APIs used

- `events.On`, `events.OnKeyDown`, `events.OnKeyUp`
- `js.Document`
- `math.Vec2`

## Example end-to-end

@include:ExampleFrame:{code:"/examples/components/input_component.go", uri:"/examples/input"}

```go
m := input.New()
m.BindMouse("pan", 1)
m.BindMouse("rotate", 2)
// ... in a frame loop
cam := m.Camera()
```

## Notes and Limitations

- Actions are identified by strings; semantics are application-defined.
- Zoom adjusts linearly and has no built-in clamping.
- Drag rectangles track only the primary mouse button.

## Related links

- [events](events)
- [math](math)
- [js](js)

