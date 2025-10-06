//go:build js && wasm

package draw

import (
	"math"

	dom "github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
)

type canvas2D struct {
	element dom.Element
	ctx     js.Value
}

// NewCanvas binds a DOM canvas element to the drawing helpers.
func NewCanvas(el dom.Element) (Canvas, bool) {
	if el.IsNull() || el.IsUndefined() {
		return Canvas{}, false
	}
	ctx := el.Call("getContext", "2d")
	if !ctx.Truthy() {
		return Canvas{}, false
	}
	return Canvas{impl: &canvas2D{element: el, ctx: ctx}}, true
}

func (c *canvas2D) valid() bool { return c.ctx.Truthy() }

func (c *canvas2D) setSize(width, height float64) {
	c.element.Set("width", int(width))
	c.element.Set("height", int(height))
}

func (c *canvas2D) drawRect(r Rect) {
	if r.fillColor != "" {
		c.ctx.Set("fillStyle", r.fillColor)
		c.ctx.Call("fillRect", r.X, r.Y, r.Width, r.Height)
	}
	if r.strokeColor != "" && r.strokeWidth > 0 {
		c.ctx.Set("lineWidth", r.strokeWidth)
		c.ctx.Set("strokeStyle", r.strokeColor)
		c.ctx.Call("strokeRect", r.X, r.Y, r.Width, r.Height)
	}
}

func (c *canvas2D) drawCircle(circle Circle) {
	c.ctx.Call("beginPath")
	c.ctx.Call("arc", circle.X, circle.Y, circle.Radius, 0, math.Pi*2, false)
	if circle.fillColor != "" {
		c.ctx.Set("fillStyle", circle.fillColor)
		c.ctx.Call("fill")
	}
	if circle.strokeColor != "" && circle.strokeWidth > 0 {
		c.ctx.Set("lineWidth", circle.strokeWidth)
		c.ctx.Set("strokeStyle", circle.strokeColor)
		c.ctx.Call("stroke")
	}
}

func (c *canvas2D) drawLine(line Line) {
	if line.strokeColor == "" || line.strokeWidth <= 0 {
		return
	}
	c.ctx.Call("beginPath")
	c.ctx.Call("moveTo", line.X1, line.Y1)
	c.ctx.Call("lineTo", line.X2, line.Y2)
	c.ctx.Set("lineWidth", line.strokeWidth)
	c.ctx.Set("strokeStyle", line.strokeColor)
	c.ctx.Call("stroke")
}
