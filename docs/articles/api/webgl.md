# webgl

Bindings for the WebGL rendering context used with WebAssembly applications.

| Function | Description |
| --- | --- |
| `NewContext(id)` | Obtain a context from a canvas element by id. |
| `NewContextFrom(canvas, type...)` | Derive a context from an existing canvas value. |
| `Context.ClearColor(r,g,b,a)` | Set the clear color. |
| `Context.Clear(mask)` | Clear buffers. |
| `Context.CreateShader(t)` | Create a shader object. |
| `Context.ShaderSource(shader, src)` | Provide GLSL source for a shader. |
| `Context.CompileShader(shader)` | Compile a shader. |
| `Context.CreateProgram()` | Create a program object. |
| `Context.AttachShader(prog, shader)` | Attach a shader to a program. |
| `Context.LinkProgram(prog)` | Link a program. |
| `Context.UseProgram(prog)` | Use a program for rendering. |
| `Context.CreateProgramFromSource(vSrc, fSrc)` | Compile shaders and link into a program. |
| `Context.BufferData(target, data, usage)` | Upload data to a buffer object. |
| `Context.BufferDataFloat32(target, data, usage)` | Upload float32 slice data to a buffer. |
| `Context.CreateBuffer()` | Create a buffer object. |
| `Context.BindBuffer(target, buffer)` | Bind a buffer to a target. |
| `Context.CreateTexture()` | Create a texture object. |
| `Context.BindTexture(target, tex)` | Bind a texture to a target. |
| `Context.ActiveTexture(texUnit)` | Select the active texture unit. |
| `Context.TexParameteri(target, pname, param)` | Set texture parameters. |
| `Context.TexImage2D(target, lvl, ifmt, w, h, border, fmt, typ, pixels)` | Upload pixel data to a texture. |
| `Context.TexImage2DFromImage(target, lvl, ifmt, fmt, typ, img)` | Upload an image/video/canvas to a texture. |
| `Context.LoadTexture2D(img)` | Create and initialize a 2D texture from an image. |
| `Context.EnableVertexAttribArray(index)` | Enable a vertex attribute. |
| `Context.VertexAttribPointer(index,size,type,normalized,stride,offset)` | Specify vertex attribute layout. |
| `Context.GetAttribLocation(prog, name)` | Retrieve attribute location. |
| `Context.GetUniformLocation(prog, name)` | Retrieve uniform location. |
| `Context.Uniform2f(loc, x, y)` | Set a vec2 uniform. |
| `Context.Uniform4f(loc, v0,v1,v2,v3)` | Set a vec4 uniform. |
| `Context.Uniform1f(loc, x)` | Set a float uniform. |
| `Context.Enable(cap)` | Enable a WebGL capability. |
| `Context.BlendFunc(sfactor, dfactor)` | Define pixel blending factors. |
| `Context.DrawArrays(mode, first, count)` | Render primitives from array data. |

## Usage

Obtain a rendering context via `webgl.NewContext` or `webgl.NewContextFrom`. The
returned `Context` value offers helpers for common tasks and a generic `Call`
method for accessing any other WebGL API.

The following component implements a simple "Snake" mini game rendered with
WebGL showcasing animated colors, blending and an interactive canvas.

@include:ExampleFrame:{code:"/examples/components/webgl_component.go", uri:"/examples/webgl"}

### Quick program creation

```go
ctx := webgl.NewContext("canvas")
prog := ctx.CreateProgramFromSource(vsSrc, fsSrc)
ctx.UseProgram(prog)
```

### Loading a texture

```go
img := dom.ByID("img")
tex := ctx.LoadTexture2D(img)
ctx.ActiveTexture(webgl.TEXTURE0)
ctx.BindTexture(webgl.TEXTURE_2D, tex)
```
