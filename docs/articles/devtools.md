# Devtools

## Development Server

The `rfw dev` command launches a file‑watching server that recompiles your project into WebAssembly on every change. It also serves the generated assets over an HTTP server so you can iterate quickly without leaving the terminal. Any files placed in a top-level `static/` directory are available at the root URL during development, and requests to `/static/*` are transparently served as `/*`. When a `host/` directory is present, `rfw dev` builds and runs the host binary from `build/host/host` so host components can be exercised locally.

Run the development server from the project directory:

```bash
rfw dev --port 8081 --debug
```

The example above enables verbose logging and profiling endpoints on port `8081`. Use this mode when you need detailed insight into build or runtime behaviour.

### Flags

- `--port` specify port
- `--host` expose to network
- `--debug` enable verbose logs and profiling endpoints (`/debug/vars`, `/debug/pprof/`)

`--host` is useful when testing on real devices; however, remember to trust the network you expose the server to. `--debug` should be left off in production builds because it serves sensitive profiling data.

## Hot Reload

Changes to Go, RTML, Markdown and plugin assets trigger automatic rebuilds and page refreshes. The watcher traverses the project tree and recompiles only the parts that changed, keeping feedback loops short.

New directories are watched automatically. On some network file systems events may not propagate; in those cases restart the server or use polling mode if available.

## Profiling

With the `--debug` flag, runtime metrics are available at dedicated endpoints. These metrics help track build frequency and memory usage while iterating on components.

- `/debug/vars` exposes counters like `rebuilds`
- `/debug/pprof/` provides pprof profiles

To inspect CPU profiles, run:

```bash
go tool pprof http://localhost:8081/debug/pprof/profile
```

Refer to the [Go pprof guide](https://pkg.go.dev/net/http/pprof) for more ways to analyse performance.

