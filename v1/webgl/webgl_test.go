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

func TestBufferDataFloat32UsesTypedArray(t *testing.T) {
	var received js.Value
	obj := js.Call("Object")

	f := js.FuncOf(func(this js.Value, args []js.Value) any {
		received = args[1]
		return nil
	})
	defer f.Release()
	obj.Set("bufferData", f)

	ctx := Context{v: obj}
	data := []float32{1, 2, 3}
	ctx.BufferDataFloat32(ARRAY_BUFFER, data, STATIC_DRAW)

	if !received.InstanceOf(js.Float32Array()) {
		t.Fatalf("expected Float32Array, got %v", received)
	}
	if n := received.Get("length").Int(); n != len(data) {
		t.Fatalf("expected length %d, got %d", len(data), n)
	}
	for i, v := range data {
		if received.Index(i).Float() != float64(v) {
			t.Fatalf("element %d mismatch", i)
		}
	}
}
