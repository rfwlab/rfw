//go:build js && wasm

// Package webgl provides minimal bindings to the WebGL rendering context.
//
// The API exposes a Context type wrapping a JavaScript WebGLRenderingContext
// allowing Go programs compiled to WebAssembly to interact with the WebGL API.
//
// Only a subset of convenience methods are implemented. The Call method can be
// used to access any other WebGL function.
package webgl

import (
	jst "syscall/js"

	js "github.com/rfwlab/rfw/v1/js"
)

// Context wraps a JavaScript WebGL rendering context.
type Context struct{ v jst.Value }

// NewContext obtains a WebGL rendering context from the canvas element with the
// provided id. It returns an empty Context if the canvas or context is not
// available.
func NewContext(canvasID string) Context {
	canvas := js.Document().Call("getElementById", canvasID)
	if canvas.IsNull() || canvas.IsUndefined() {
		return Context{}
	}
	ctx := canvas.Call("getContext", "webgl")
	return Context{v: ctx}
}

// NewContextFrom obtains a WebGL rendering context from an existing canvas
// element value. If ctxType is empty, "webgl" is used.
func NewContextFrom(canvas jst.Value, ctxType ...string) Context {
	t := "webgl"
	if len(ctxType) > 0 {
		t = ctxType[0]
	}
	return Context{v: canvas.Call("getContext", t)}
}

// Value returns the underlying JavaScript value of the context.
func (c Context) Value() jst.Value { return c.v }

// Call invokes a WebGL function on the context by name. It can be used to
// access WebGL methods not covered by convenience wrappers.
func (c Context) Call(name string, args ...any) jst.Value { return c.v.Call(name, args...) }

// Get retrieves a property from the context.
func (c Context) Get(name string) jst.Value { return c.v.Get(name) }

// ClearColor sets the clear color.
func (c Context) ClearColor(r, g, b, a float32) { c.v.Call("clearColor", r, g, b, a) }

// Clear clears buffers specified by mask.
func (c Context) Clear(mask int) { c.v.Call("clear", mask) }

// CreateShader creates a shader of the given type.
func (c Context) CreateShader(t int) jst.Value { return c.v.Call("createShader", t) }

// ShaderSource sets the source code of the shader.
func (c Context) ShaderSource(shader jst.Value, src string) { c.v.Call("shaderSource", shader, src) }

// CompileShader compiles the given shader.
func (c Context) CompileShader(shader jst.Value) { c.v.Call("compileShader", shader) }

// CreateProgram creates a new program object.
func (c Context) CreateProgram() jst.Value { return c.v.Call("createProgram") }

// AttachShader attaches a shader to a program.
func (c Context) AttachShader(program, shader jst.Value) { c.v.Call("attachShader", program, shader) }

// LinkProgram links a program.
func (c Context) LinkProgram(program jst.Value) { c.v.Call("linkProgram", program) }

// UseProgram sets the active program.
func (c Context) UseProgram(program jst.Value) { c.v.Call("useProgram", program) }

// BufferData uploads data to a buffer object.
func (c Context) BufferData(target int, data jst.Value, usage int) {
	c.v.Call("bufferData", target, data, usage)
}

// BufferDataFloat32 uploads a float32 slice to a buffer object by creating a
// JavaScript Float32Array and passing it to BufferData.
func (c Context) BufferDataFloat32(target int, data []float32, usage int) {
	arr := js.Get("Float32Array").New(len(data))
	for i, v := range data {
		arr.SetIndex(i, v)
	}
	c.BufferData(target, arr, usage)
}

// CreateBuffer creates a new buffer object.
func (c Context) CreateBuffer() jst.Value { return c.v.Call("createBuffer") }

// BindBuffer binds a buffer to a target.
func (c Context) BindBuffer(target int, buffer jst.Value) {
	c.v.Call("bindBuffer", target, buffer)
}

// EnableVertexAttribArray enables a vertex attribute array at the given index.
func (c Context) EnableVertexAttribArray(index int) {
	c.v.Call("enableVertexAttribArray", index)
}

// VertexAttribPointer defines an array of generic vertex attribute data.
func (c Context) VertexAttribPointer(index, size, typ int, normalized bool, stride, offset int) {
	c.v.Call("vertexAttribPointer", index, size, typ, normalized, stride, offset)
}

// GetAttribLocation returns the location of an attribute variable in a program.
func (c Context) GetAttribLocation(program jst.Value, name string) int {
	return c.v.Call("getAttribLocation", program, name).Int()
}

// GetUniformLocation returns the location of a uniform variable.
func (c Context) GetUniformLocation(program jst.Value, name string) jst.Value {
	return c.v.Call("getUniformLocation", program, name)
}

// Uniform2f sets a vec2 uniform value.
func (c Context) Uniform2f(loc jst.Value, x, y float32) { c.v.Call("uniform2f", loc, x, y) }

// Uniform4f sets a vec4 uniform value.
func (c Context) Uniform4f(loc jst.Value, v0, v1, v2, v3 float32) {
	c.v.Call("uniform4f", loc, v0, v1, v2, v3)
}

// Uniform1f sets a float uniform value.
func (c Context) Uniform1f(loc jst.Value, x float32) { c.v.Call("uniform1f", loc, x) }

// Enable enables a specific WebGL capability.
func (c Context) Enable(cap int) { c.v.Call("enable", cap) }

// BlendFunc defines the pixel blending factors.
func (c Context) BlendFunc(sfactor, dfactor int) { c.v.Call("blendFunc", sfactor, dfactor) }

// DrawArrays renders primitives from array data.
func (c Context) DrawArrays(mode, first, count int) { c.v.Call("drawArrays", mode, first, count) }

// Constants related to WebGL operations.
const (
	COLOR_BUFFER_BIT   = 0x4000
	DEPTH_BUFFER_BIT   = 0x0100
	STENCIL_BUFFER_BIT = 0x0400

	TRIANGLES    = 0x0004
	ARRAY_BUFFER = 0x8892
	STATIC_DRAW  = 0x88E4

	FLOAT = 0x1406

	VERTEX_SHADER   = 0x8B31
	FRAGMENT_SHADER = 0x8B30

	BLEND     = 0x0BE2
	SRC_ALPHA = 0x0302
	ONE       = 1
)
