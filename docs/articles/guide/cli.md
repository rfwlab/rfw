# CLI Guide

The `rfw` command-line interface manages project scaffolding, development and builds.

## `rfw init <module>`

Creates a new RFW project by cloning a template and initializing a Go module.

```bash
rfw init github.com/username/hello-rfw
```

Expected output:

```text
Project 'hello-rfw' initialized successfully.
Project initialized
```

The directory `hello-rfw/` will contain `main.go`, a `components/` folder with `.rtml` templates, a `host/` directory with server components, and an `rfw.json` manifest preconfigured for SSC builds.

## `rfw dev [--port --host --debug]`

Starts the development server, compiles the project to WebAssembly and rebuilds on file changes. Files placed in `static/` are served at the root path during development, and requests to `/static/*` are transparently mapped to `/*` to mirror production builds. When a `host/` directory exists, the command builds and runs the host binary from `build/host/host`.

Flags:

- `--port` choose the port (default `8080`)
- `--host` expose the server to the local network
- `--debug` enable verbose logs and profiling endpoints (`/debug/vars`, `/debug/pprof/`)

Example:

```bash
rfw dev --port 3000 --host --debug
```

Expected output:

```text
rfw v0.1.3

➜ Local: http://localhost:3000/
➜ Network: http://192.168.1.10:3000/
➜ Press h + enter to show help
```

## `rfw build`

Compiles the current project into an optimized Wasm bundle ready for deployment.
Artifacts are written to `build/client/` and a host binary is produced in `build/host/` to serve the client and optional host components. Any files placed in `static/` are copied into `build/static/` and served at the root path.

```bash
rfw build
```

Expected output:

```text
Build completed
```
