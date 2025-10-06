# game draw

Declarative helpers for issuing 2D canvas drawing commands.

| Function | Description |
| --- | --- |
| `NewCanvas(el dom.Element) (Canvas, bool)` | Acquire a 2D rendering surface from a `<canvas>` element. |
| `(Canvas).Valid() bool` | Report whether the canvas has an active 2D context. |
| `(Canvas).SetSize(width, height float64)` | Update the backing canvas width and height in pixels. |
| `(Canvas).Draw(cmds ...Command)` | Execute the provided drawing commands, skipping `nil` entries. |
| `Rectangle(x, y, width, height float64) *Rect` | Create a rectangle command for the given bounds. |
| `(*Rect).Fill(color string) *Rect` | Set the rectangle fill color. |
| `(*Rect).Stroke(color string, width float64) *Rect` | Set the rectangle stroke color and width. |
| `Disc(x, y, radius float64) *Circle` | Create a circle command centered at the provided coordinates. |
| `(*Circle).Fill(color string) *Circle` | Set the circle fill color. |
| `(*Circle).Stroke(color string, width float64) *Circle` | Set the circle stroke color and width. |
| `Segment(x1, y1, x2, y2 float64) *Line` | Create a line segment command between the two points. |
| `(*Line).Stroke(color string, width float64) *Line` | Set the line stroke color and width. |
