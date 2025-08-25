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
| `Context.BufferData(target, data, usage)` | Upload data to a buffer object. |

## Usage

Obtain a rendering context via `webgl.NewContext` or `webgl.NewContextFrom`. The
returned `Context` value offers helpers for common tasks and a generic `Call`
method for accessing any other WebGL API.

The following component clears a canvas to red using WebGL.

@include:ExampleFrame:{code:"/examples/components/webgl_component.go", uri:"/examples/webgl"}
