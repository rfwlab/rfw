# Quick Start

This guide walks through creating and running your first RFW application using the command‑line interface. All commands assume a Go environment configured for WebAssembly.

## Install the CLI

```bash
go install github.com/rfwlab/rfw/cmd/rfw@latest
```

Verify installation with `rfw -h`.

## Scaffold a Project

Initialize a new module. The command creates a Go module with a sample component and entry point:

```bash
rfw init github.com/username/hello-rfw
cd hello-rfw
```

The generated project contains `main.go`, a `components` folder, and matching `.rtml` templates.

## Run the Development Server

```bash
rfw dev
```

`rfw dev` compiles the Go sources to `app.wasm`, serves an HTML shell, and reloads the page when files change.

The default `index.html` mounts the Wasm module with plain JavaScript:

```html
<script src="/wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch('/app.wasm'), go.importObject).then((result) => {
    go.run(result.instance);
  });
</script>
```

A global `rfw` object will be added in upcoming releases to expose high‑level APIs to JavaScript. Until then, interaction with the framework in the browser should stick to plain JS.

## Build for Production

```bash
rfw build
```

The build command outputs an optimized Wasm file and accompanying assets ready to deploy.

With these basics in place, dive deeper into the framework starting with [Creating an Application](../essentials/creating-application.md).
