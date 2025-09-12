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
# or skip go mod tidy using the --skip-tidy flag:
# rfw init --skip-tidy github.com/username/hello-rfw
```

The generated project contains `main.go`, a `components/` folder, matching `.rtml` templates, a `host/` directory with its own `components/`, and an `rfw.json` manifest enabling SSC builds.

## Run the Development Server

```bash
rfw dev
```

`rfw dev` compiles the Go sources to `app.wasm`, serves an HTML shell, reloads the page when files change, and serves files from a top-level `static/` directory at the root path. Requests to `/static/*` resolve to the same files so URLs stay unchanged. When a `host/` directory is present, the command also builds and runs the host binary from `build/host/host`.

The server listens on port `8080` by default. Use the `--port` flag or set the
`RFW_PORT` environment variable to change it:

```bash
RFW_PORT=3000 rfw dev
```

Set the `RFW_LOG_LEVEL` environment variable to control log verbosity
(`debug`, `info`, `warn`, `error`):

```bash
RFW_LOG_LEVEL=debug rfw dev
```

The default `index.html` mounts the Wasm module with plain JavaScript:

```html
<script src="/wasm_exec.js"></script>
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load('/app.wasm', { go });
</script>
```

A global `rfw` object will be added in upcoming releases to expose high‑level APIs to JavaScript. Until then, interaction with the framework in the browser should stick to plain JS.

## Build for Production

```bash
rfw build
```

The build command writes the Wasm bundle and supporting files to `build/client/`. Files placed in a top-level `static/` directory are copied into `build/static/` and served at the root path.
The companion host binary, used to serve the client and optional host components, is placed under `build/host/`.

With these basics in place, dive deeper into the framework starting with [Creating an Application](../essentials/creating-application).
