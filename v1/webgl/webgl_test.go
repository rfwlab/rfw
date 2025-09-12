//go:build js && wasm

package webgl

import (
	"testing"

	js "github.com/rfwlab/rfw/v1/js"
)

func TestWrappersInvokeContext(t *testing.T) {
	called := map[string]bool{}
	obj := js.Call("Object")

	fCreate := js.FuncOf(func(this js.Value, args []js.Value) any {
		called["createVertexArray"] = true
		return js.Call("Object")
	})
	defer fCreate.Release()
	obj.Set("createVertexArray", fCreate)

	fBind := js.FuncOf(func(this js.Value, args []js.Value) any {
		called["bindVertexArray"] = true
		return nil
	})
	defer fBind.Release()
	obj.Set("bindVertexArray", fBind)

	fDraw := js.FuncOf(func(this js.Value, args []js.Value) any {
		called["drawElements"] = true
		return nil
	})
	defer fDraw.Release()
	obj.Set("drawElements", fDraw)

	fViewport := js.FuncOf(func(this js.Value, args []js.Value) any {
		called["viewport"] = true
		return nil
	})
	defer fViewport.Release()
	obj.Set("viewport", fViewport)

	fDepth := js.FuncOf(func(this js.Value, args []js.Value) any {
		called["depthFunc"] = true
		return nil
	})
	defer fDepth.Release()
	obj.Set("depthFunc", fDepth)

	ctx := Context{v: obj}
	vao := ctx.CreateVertexArray()
	ctx.BindVertexArray(vao)
	ctx.DrawElements(TRIANGLES, 3, UNSIGNED_SHORT, 0)
	ctx.Viewport(0, 0, 1, 1)
	ctx.DepthFunc(LEQUAL)

	for name, ok := range called {
		if !ok {
			t.Errorf("%s was not called", name)
		}
	}
}
