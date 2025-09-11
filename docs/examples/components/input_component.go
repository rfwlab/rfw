//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/input"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/input_component.rtml
var inputComponentTpl []byte

// NewInputComponent demonstrates the input manager and camera helpers.
func NewInputComponent() *core.HTMLComponent {
	c := core.NewComponent("InputComponent", inputComponentTpl, nil)

	m := input.New()
	m.BindMouse("pan", 1)
	m.BindMouse("rotate", 2)

	var loop js.Func
	loop = js.FuncOf(func(this js.Value, args []js.Value) any {
		cam := m.Camera()
		dom.SetText(dom.ByID("camera"), fmt.Sprintf("pos(%.0f,%.0f) zoom %.2f rot %.0f", cam.Position.X, cam.Position.Y, cam.Zoom, cam.Rotation))
		if start, end, dragging := m.DragRect(); dragging {
			dom.SetText(dom.ByID("drag"), fmt.Sprintf("drag (%.0f,%.0f)-(%.0f,%.0f)", start.X, start.Y, end.X, end.Y))
		} else {
			dom.SetText(dom.ByID("drag"), "")
		}
		js.RequestAnimationFrame(loop)
		return nil
	})
	js.RequestAnimationFrame(loop)

	return c
}
