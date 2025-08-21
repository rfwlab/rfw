# Server-Side Rendering

> Be careful, even if the whole rfw project is still in Alpha, this specific feature must be considered more than experimental.

The `v1/ssr` package converts RTML templates into HTML on the server at request time, returning fully rendered pages before the browser loads the Wasm bundle.

## Flow

1. The project manifest (`rfw.json`) sets `build.type` to `ssr`.
2. `cmd/rfw/build` compiles the Wasm bundle to `dist/client/app.wasm` and copies `wasm_exec.js` into the same directory.
3. `cmd/rfw/server` renders `index.rtml` on each request using `ssr.RenderFile`, wrapping the result in `<div id="app" data-hydrate>`.
4. The server responds with the rendered HTML and serves the files from `/dist/client/` for client hydration.

## API

```go
html, err := ssr.RenderFile("index.rtml", props)
```

`RenderFile` loads an RTML template from disk and returns the generated HTML string. A `props` map can replace `{{key}}` placeholders in the template.

## Example

Assume a file `hello.rtml`:

```html
<root><p>Hello {{name}}</p></root>
```

Render it on the server with:

```go
html, _ := ssr.RenderFile("hello.rtml", map[string]any{"name": "rfw"})
fmt.Println(html) // <root><p>Hello rfw</p></root>
```

The server should embed the Wasm loader so the client can hydrate the markup:

```html
<script src="/dist/client/wasm_exec.js"></script>
<script>
const go = new Go();
WebAssembly.instantiateStreaming( fetch("/dist/client/app.wasm?" + Date.now()), go.importObject, ).then((result) => { go.run(result.instance); });});
</script>
```

This setup delivers an initial server-side render followed by client-side hydration of the Wasm bundle. See `examples/ssr_example` for a working server.
