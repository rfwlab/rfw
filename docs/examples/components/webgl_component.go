//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	webgl "github.com/rfwlab/rfw/v1/webgl"
)

//go:embed templates/webgl_component.rtml
var webglComponentTpl []byte

// NewWebGLComponent returns a component demonstrating WebGL usage.
func NewWebGLComponent() *core.HTMLComponent {
	c := core.NewComponent("WebGLComponent", webglComponentTpl, nil)
	dom.RegisterHandlerFunc("webglDraw", func() {
		ctx := webgl.NewContext("glcanvas")
		ctx.ClearColor(1, 0, 0, 1)
		ctx.Clear(webgl.COLOR_BUFFER_BIT)
	})
	return c
}
