# CLI Guide

The `rfw` CLI manages scaffolding, development, and builds for your projects.

## `rfw init <module>`

Create a new rfw v2 project. Clones a template, sets up a Go module, and prepares project directories.

**Flags**

* `--skip-tidy` skip running `go mod tidy`

```bash
rfw init github.com/username/hello-rfw
```

The new directory includes:

* `main.go`, entry point with `composition.RegisterFS()` and `router.Page()`
* `pages/` with `.go` component files and `templates/` for RTML
* `components/` for shared components like layouts
* `host/` for SSC server components
* `rfw.json` manifest (preconfigured for SSC builds)

## `rfw dev [--port --host]`

Start the development server. Compiles the project to WebAssembly, watches files, and rebuilds on changes.

If a `host/` directory exists, the host binary is built and run from `build/host/host`.

**Flags**

* `--port` choose port (default `8080`)
* `--host` expose server to local network

Example:

```bash
rfw dev --port 3000 --host
```

## `rfw build`

Build the project for production. Generates an optimized Wasm bundle and host binary.

Artifacts:

* `build/client/`, client bundle (`app.wasm` and Brotli variant `app.wasm.br`)
* `build/host/`, host binary
* `build/static/`, copied static files

```bash
rfw build
```