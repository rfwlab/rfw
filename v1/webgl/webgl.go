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
	dom "github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
)

// Context wraps a JavaScript WebGL rendering context.
type Context struct{ v js.Value }

// NewContext obtains a WebGL rendering context from the canvas element with the
// provided id. It returns an empty Context if the canvas or context is not
// available.
func NewContext(canvasID string) Context {
	canvas := dom.Doc().ByID(canvasID)
	if canvas.IsNull() || canvas.IsUndefined() {
		return Context{}
	}
	ctx := canvas.Call("getContext", "webgl")
	return Context{v: ctx}
}

// NewContextFrom obtains a WebGL rendering context from an existing canvas
// element value. If ctxType is empty, "webgl" is used.
func NewContextFrom(canvas js.Value, ctxType ...string) Context {
	t := "webgl"
	if len(ctxType) > 0 {
		t = ctxType[0]
	}
	return Context{v: canvas.Call("getContext", t)}
}

// Value returns the underlying JavaScript value of the context.
func (c Context) Value() js.Value { return c.v }

// Call invokes a WebGL function on the context by name. It can be used to
// access WebGL methods not covered by convenience wrappers.
func (c Context) Call(name string, args ...any) js.Value { return c.v.Call(name, args...) }

// Get retrieves a property from the context.
func (c Context) Get(name string) js.Value { return c.v.Get(name) }

// ClearColor sets the clear color.
func (c Context) ClearColor(r, g, b, a float32) { c.v.Call("clearColor", r, g, b, a) }

// Clear clears buffers specified by mask.
func (c Context) Clear(mask int) { c.v.Call("clear", mask) }

// CreateShader creates a shader of the given type.
func (c Context) CreateShader(t int) js.Value { return c.v.Call("createShader", t) }

// ShaderSource sets the source code of the shader.
func (c Context) ShaderSource(shader js.Value, src string) { c.v.Call("shaderSource", shader, src) }

// CompileShader compiles the given shader.
func (c Context) CompileShader(shader js.Value) { c.v.Call("compileShader", shader) }

// CreateProgram creates a new program object.
func (c Context) CreateProgram() js.Value { return c.v.Call("createProgram") }

// AttachShader attaches a shader to a program.
func (c Context) AttachShader(program, shader js.Value) { c.v.Call("attachShader", program, shader) }

// LinkProgram links a program.
func (c Context) LinkProgram(program js.Value) { c.v.Call("linkProgram", program) }

// UseProgram sets the active program.
func (c Context) UseProgram(program js.Value) { c.v.Call("useProgram", program) }

// CreateProgramFromSource compiles the provided vertex and fragment shader
// sources, links them into a program and returns it.
func (c Context) CreateProgramFromSource(vertexSrc, fragmentSrc string) js.Value {
	vs := c.CreateShader(VERTEX_SHADER)
	c.ShaderSource(vs, vertexSrc)
	c.CompileShader(vs)
	fs := c.CreateShader(FRAGMENT_SHADER)
	c.ShaderSource(fs, fragmentSrc)
	c.CompileShader(fs)
	prog := c.CreateProgram()
	c.AttachShader(prog, vs)
	c.AttachShader(prog, fs)
	c.LinkProgram(prog)
	return prog
}

// BufferData uploads data to a buffer object.
func (c Context) BufferData(target int, data js.Value, usage int) {
	c.v.Call("bufferData", target, data, usage)
}

// BufferDataFloat32 uploads a float32 slice to a buffer object by creating a
// JavaScript Float32Array and passing it to BufferData.
func (c Context) BufferDataFloat32(target int, data []float32, usage int) {
	arr := js.Float32Array().New(len(data))
	for i, v := range data {
		arr.SetIndex(i, v)
	}
	c.BufferData(target, arr, usage)
}

// CreateBuffer creates a new buffer object.
func (c Context) CreateBuffer() js.Value { return c.v.Call("createBuffer") }

// BindBuffer binds a buffer to a target.
func (c Context) BindBuffer(target int, buffer js.Value) {
	c.v.Call("bindBuffer", target, buffer)
}

// CreateVertexArray creates a new vertex array object.
func (c Context) CreateVertexArray() js.Value { return c.v.Call("createVertexArray") }

// BindVertexArray binds a vertex array object.
func (c Context) BindVertexArray(vao js.Value) { c.v.Call("bindVertexArray", vao) }

// CreateTexture creates a new texture object.
func (c Context) CreateTexture() js.Value { return c.v.Call("createTexture") }

// BindTexture binds a texture to a target.
func (c Context) BindTexture(target int, texture js.Value) {
	c.v.Call("bindTexture", target, texture)
}

// ActiveTexture selects the active texture unit.
func (c Context) ActiveTexture(texture int) { c.v.Call("activeTexture", texture) }

// TexParameteri sets texture parameters.
func (c Context) TexParameteri(target, pname, param int) {
	c.v.Call("texParameteri", target, pname, param)
}

// TexImage2D uploads pixel data to a 2D texture.
func (c Context) TexImage2D(target, level, internalformat, width, height, border, format, typ int, pixels js.Value) {
	c.v.Call("texImage2D", target, level, internalformat, width, height, border, format, typ, pixels)
}

// TexImage2DFromImage uploads an image, video or canvas source to a texture.
func (c Context) TexImage2DFromImage(target, level, internalformat, format, typ int, img js.Value) {
	c.v.Call("texImage2D", target, level, internalformat, format, typ, img)
}

// LoadTexture2D creates a texture and initializes it from an image source using
// linear filtering. The created texture is returned.
func (c Context) LoadTexture2D(img js.Value) js.Value {
	tex := c.CreateTexture()
	c.BindTexture(TEXTURE_2D, tex)
	c.TexImage2DFromImage(TEXTURE_2D, 0, RGBA, RGBA, UNSIGNED_BYTE, img)
	c.TexParameteri(TEXTURE_2D, TEXTURE_MIN_FILTER, LINEAR)
	c.TexParameteri(TEXTURE_2D, TEXTURE_MAG_FILTER, LINEAR)
	return tex
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
func (c Context) GetAttribLocation(program js.Value, name string) int {
	return c.v.Call("getAttribLocation", program, name).Int()
}

// GetUniformLocation returns the location of a uniform variable.
func (c Context) GetUniformLocation(program js.Value, name string) js.Value {
	return c.v.Call("getUniformLocation", program, name)
}

// Uniform2f sets a vec2 uniform value.
func (c Context) Uniform2f(loc js.Value, x, y float32) { c.v.Call("uniform2f", loc, x, y) }

// Uniform4f sets a vec4 uniform value.
func (c Context) Uniform4f(loc js.Value, v0, v1, v2, v3 float32) {
	c.v.Call("uniform4f", loc, v0, v1, v2, v3)
}

// Uniform1f sets a float uniform value.
func (c Context) Uniform1f(loc js.Value, x float32) { c.v.Call("uniform1f", loc, x) }

// Enable enables a specific WebGL capability.
func (c Context) Enable(cap int) { c.v.Call("enable", cap) }

// BlendFunc defines the pixel blending factors.
func (c Context) BlendFunc(sfactor, dfactor int) { c.v.Call("blendFunc", sfactor, dfactor) }

// DrawArrays renders primitives from array data.
func (c Context) DrawArrays(mode, first, count int) { c.v.Call("drawArrays", mode, first, count) }

// DrawElements renders primitives using an index buffer.
func (c Context) DrawElements(mode, count, typ, offset int) {
	c.v.Call("drawElements", mode, count, typ, offset)
}

// Viewport sets the viewport dimensions.
func (c Context) Viewport(x, y, width, height int) { c.v.Call("viewport", x, y, width, height) }

// DepthFunc specifies the depth comparison function.
func (c Context) DepthFunc(fn int) { c.v.Call("depthFunc", fn) }

// Constants related to WebGL operations.
const (
	COLOR_BUFFER_BIT   = 0x4000
	DEPTH_BUFFER_BIT   = 0x0100
	STENCIL_BUFFER_BIT = 0x0400

	TRIANGLES            = 0x0004
	ARRAY_BUFFER         = 0x8892
	STATIC_DRAW          = 0x88E4
	ELEMENT_ARRAY_BUFFER = 0x8893

	FLOAT = 0x1406

	VERTEX_SHADER   = 0x8B31
	FRAGMENT_SHADER = 0x8B30

	BLEND      = 0x0BE2
	DEPTH_TEST = 0x0B71
	SRC_ALPHA  = 0x0302
	ONE        = 1

	TEXTURE_2D         = 0x0DE1
	TEXTURE0           = 0x84C0
	RGBA               = 0x1908
	UNSIGNED_BYTE      = 0x1401
	UNSIGNED_SHORT     = 0x1403
	TEXTURE_MIN_FILTER = 0x2801
	TEXTURE_MAG_FILTER = 0x2800
	LINEAR             = 0x2601

	// Depth function comparisons.
	NEVER    = 0x0200
	LESS     = 0x0201
	EQUAL    = 0x0202
	LEQUAL   = 0x0203
	GREATER  = 0x0204
	NOTEQUAL = 0x0205
	GEQUAL   = 0x0206
	ALWAYS   = 0x0207
)
