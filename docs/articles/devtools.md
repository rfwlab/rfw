# Devtools

## Enabling the overlay

Set `RFW_DEVTOOLS=1` before running any CLI command to include the debugging scripts in the build. This environment variable compiles the DevTools package and exposes runtime hooks.

```bash
RFW_DEVTOOLS=1 rfw dev --debug
```

Avoid enabling it for production builds where the overlay is unnecessary.

## Runtime error overlay

### Why
Surface uncaught panics and JavaScript errors directly in the browser.

### When to use
Available in development when `RFW_DEVTOOLS=1` is set.

### How
1. Trigger an error, e.g. `panic("boom")`.
2. The overlay shows the message and stack trace.
3. Use the arrow buttons to navigate multiple errors or reload the page.

### API
The overlay hooks `window` error events through the [`js` package](../api/js) and requires no application code.

### Example
```go
func broken() {
    panic("boom")
}
```

@include:ExampleFrame:{code:"/examples/components/runtime_error_component.go", uri:"/examples/runtime-error"}

### Notes
- Active only in debug builds.
- Closing the overlay clears captured errors.
- Stores at most 15 errors per page; additional ones are ignored.

### Related links
- [js package](../api/js)

## Development Server

The `rfw dev` command launches a file‑watching server that recompiles your project into WebAssembly on every change. It also serves the generated assets over an HTTP server so you can iterate quickly without leaving the terminal. Any files placed in a top-level `static/` directory are available at the root URL during development, and requests to `/static/*` are transparently served as `/*`. When a `host/` directory is present, `rfw dev` builds and runs the host binary from `build/host/host` so host components can be exercised locally.

By default the server listens on port `8080`. Override it with the `--port`
flag or the `RFW_PORT` environment variable.

Run the development server from the project directory:

```bash
rfw dev --port 8081 --debug
```

The example above enables verbose logging and profiling endpoints on port `8081`. These routes are exposed by both the development server and any host binary, so `/debug/vars` and `/debug/pprof/` remain available when server‑side components are active. Use this mode when you need detailed insight into build or runtime behaviour.

### Flags

- `--port` specify port
- `--host` expose to network
- `--debug` enable verbose logs and profiling endpoints (`/debug/vars`, `/debug/pprof/`)

The port can also be set with the `RFW_PORT` environment variable.

Log verbosity can be tuned with the `RFW_LOG_LEVEL` environment variable
(`debug`, `info`, `warn`, `error`).

`--host` is useful when testing on real devices; however, remember to trust the network you expose the server to. `--debug` should be left off in production builds because it serves sensitive profiling data.

## Hot Reload

Changes to Go, RTML, Markdown and plugin assets trigger automatic rebuilds and page refreshes. The watcher traverses the project tree and recompiles only the parts that changed, keeping feedback loops short.

New directories are watched automatically. On some network file systems events may not propagate; in those cases restart the server or use polling mode if available.

## Profiling

 With the `--debug` flag, runtime metrics are available at dedicated endpoints. These metrics help track build frequency and memory usage while iterating on components. The DevTools overlay surfaces this data under two additional tabs:

- **Vars** presents the JSON from `/debug/vars` as a searchable tree.
- **Pprof** lists profiles under `/debug/pprof/`. Text-based profiles render inline, while binary profiles offer a download link for inspection with external tools.

To inspect CPU profiles, run:

```bash
go tool pprof http://localhost:8081/debug/pprof/profile
```

Refer to the [Go pprof guide](https://pkg.go.dev/net/http/pprof) for more ways to analyse performance.

## Component tree API

The debug overlay mirrors the active component hierarchy by collecting lifecycle events.

### When to use

Enable the development server with `--debug` to inspect the tree during local development.

### How it works

The runtime registers hooks via `core.RegisterLifecycle` and `RegisterRouter` so mounts, unmounts and navigation rebuild the structure. The current tree is exposed to JavaScript as `RFW_DEVTOOLS_TREE()` which returns JSON.

### Example

```js
// Dump the active component hierarchy
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_TREE()));
```

### Refreshing

Use the refresh button next to the component filter to rebuild the list. The overlay
also refreshes automatically on navigation.

```js
// Manual refresh from the console
globalThis.RFW_DEVTOOLS_REFRESH();
```

### Limitations

This function exists only in debug builds and should not be relied upon in production code.

Components added dynamically must be registered with [`AddDependency`](../api/core#dependency-injection) or use `rt-is` so the overlay can display them. This includes markdown-driven pages that mount components after parsing to display embedded examples.

### Related links

- [core lifecycle hooks](../api/core#lifecycle-hooks)
- [router](../api/router#usage)

## Store inspection

Inspect global state stores from the **Store** tab. Each module, store and key
is listed in a tree view similar to the component browser.

### When to use

Use this tab during development to observe mutations without instrumenting
application code.

### How it works

The DevTools plugin registers a store hook via `core.RegisterStore`, capturing
updates and exposing the current snapshot through
`RFW_DEVTOOLS_STORES()`.

### Example

```js
// Dump all registered stores
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_STORES()));
```

### Limitations

Only stores created in debug builds are reported. Avoid using this API in
production.

### Related links

- [state StoreHook](../api/state#storehook)
- [Stores vs signals](../guide/store-vs-signals)

## Signal inspection

The **Signals** tab lists reactive signals and their current values.

### When to use

Use it to debug local reactive state while iterating on components.

### How it works

Signals register with the DevTools runtime when created. The snapshot is exposed
through `RFW_DEVTOOLS_SIGNALS()` and refreshes on every `Set`.

### Example

```js
// List all tracked signals
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_SIGNALS()));
```

### Limitations

Signals are anonymous; the tab identifies them by numerical ID. The feature is
available only in debug builds.

### Related links

- [Signals & Effects](../essentials/signals-and-effects)

## Plugin inspection

The **Plugins** tab lists build plugins that are active along with their configuration.

### When to use

Use it during development to verify which plugins ran and the parameters they received.

### How it works

The CLI exposes initialized plugins and their configuration directly from the plugin registry instead of reading `rfw.json`.

### Example

```js
// Inspect active plugins
console.log(globalThis.RFW_DEVTOOLS_PLUGINS());
```

### Limitations

Only successfully configured plugins appear in this tab. The feature is available only in debug builds.

### Related links

- [plugins](../api/plugins)
- [Manifest plugins](../guide/manifest#plugins)

## Network inspector

Monitor HTTP requests from the **Network** tab.

### When to use

Use it during development to see request URLs, status codes and timings without instrumenting application code.

### How it works

The runtime wraps [`http.FetchJSON`](../api/http) and forwards request events through a hook registered with `http.RegisterHTTPHook`. DevTools records completed requests and displays their status and duration.

### Example

```go
import (
    "log"
    "time"
)

http.RegisterHTTPHook(func(start bool, url string, status int, d time.Duration) {
    if !start {
        log.Printf("%s -> %d in %v", url, status, d)
    }
})
```

@include:ExampleFrame:{code:"/examples/components/fetchjson_component.go", uri:"/examples/fetchjson"}

### Limitations

Only requests made via `FetchJSON` are tracked. The feature is available in debug builds when `RFW_DEVTOOLS=1`.

### Related links

- [http helpers](../api/http)

