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

The directory `hello-rfw/` will contain `main.go`, a `components/` folder and matching `.rtml` templates.

## `rfw dev [--port --host --debug]`

Starts the development server, compiles the project to WebAssembly and rebuilds on file changes.

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

```bash
rfw build
```

Expected output:

```text
Build completed
```
