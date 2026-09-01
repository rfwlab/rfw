//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"

	core "github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/input"
	js "github.com/rfwlab/rfw/v2/js"
)

//go:embed templates/input_component.rtml
var inputComponentTpl []byte

type InputComponent struct {
	*core.HTMLComponent
	manager *input.Manager
	active  bool
	loop    js.Func
}

// NewInputComponent demonstrates the input manager and camera helpers.
func NewInputComponent() *core.HTMLComponent {
	c := &InputComponent{manager: input.New()}
	c.manager.BindMouse("pan", 1)
	c.manager.BindMouse("rotate", 2)
	c.HTMLComponent = core.NewComponentWith("InputComponent", inputComponentTpl, nil, c)
	return c.HTMLComponent
}

func (c *InputComponent) OnMount() {
	c.active = true
	camEl := c.GetRef("camera")
	dragEl := c.GetRef("drag")
	c.loop = js.FuncOf(func(this js.Value, args []js.Value) any {
		if !c.active || !c.IsMounted() {
			c.loop.Release()
			return nil
		}
		cam := c.manager.Camera()
		camEl.SetText(fmt.Sprintf("pos(%.0f,%.0f) zoom %.2f rot %.0f", cam.Position.X, cam.Position.Y, cam.Zoom, cam.Rotation))
		if start, end, dragging := c.manager.DragRect(); dragging {
			dragEl.SetText(fmt.Sprintf("drag (%.0f,%.0f)-(%.0f,%.0f)", start.X, start.Y, end.X, end.Y))
		} else {
			dragEl.SetText("")
		}
		js.RequestAnimationFrame(c.loop)
		return nil
	})
	js.RequestAnimationFrame(c.loop)
}

func (c *InputComponent) OnUnmount() {
	c.active = false
}
