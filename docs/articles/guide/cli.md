# CLI Guide

The `rfw` CLI manages scaffolding, development, and builds for your projects.

## `rfw init <module>`

Create a new rfw project. This clones a template, sets up a Go module, and prepares project directories.

**Flags**

* `--skip-tidy` skip running `go mod tidy`

```bash
rfw init github.com/username/hello-rfw
# or
rfw init --skip-tidy github.com/username/hello-rfw
```

After running, you’ll see:

```text
Project 'hello-rfw' initialized successfully.
Project initialized
```

The new directory includes:

* `main.go`
* `components/` with `.rtml` templates
* `host/` for server components
* `rfw.json` manifest (preconfigured for SSC builds)

## `rfw dev [--port --host --debug]`

Start the development server. It compiles the project to WebAssembly, watches files, and rebuilds on changes. Files in `static/` are served at the root. Requests to `/static/*` map to `/*` to mirror production.

If a `host/` directory exists, the host binary is built and run from `build/host/host`.

**Flags**

* `--port` choose port (default `8080`)
* `--host` expose server to local network
* `--debug` enable verbose logs and profiling (`/debug/vars`, `/debug/pprof/`)

When `--debug` is enabled, the bundler plugin is skipped so assets stay unminified. Use `rfw build` for optimized output.

Example:

```bash
rfw dev --port 3000 --host --debug
```

Output:

```text
rfw v0.2.0-beta.1

➜ Local: http://localhost:3000/
➜ Network: http://192.168.1.10:3000/
➜ Press h + enter to show help
```

## `rfw build`

Build the project for production. Generates an optimized Wasm bundle and host binary.

Artifacts:

* `build/client/` – client bundle
* `build/host/` – host binary
* `build/static/` – copied static files

```bash
rfw build
```

Output:

```text
Build completed
```
