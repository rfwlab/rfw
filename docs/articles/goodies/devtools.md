# DevTools

## Enabling the overlay

Set `RFW_DEVTOOLS=1` before running any CLI command to include the debugging overlay. This compiles the DevTools package and exposes runtime hooks.

```bash
RFW_DEVTOOLS=1 rfw dev --debug
```

Avoid enabling it in production builds.

## Runtime error overlay

Shows uncaught panics and JavaScript errors directly in the browser.

**Usage**

1. Trigger an error, e.g. `panic("boom")`.
2. The overlay displays the message and stack trace.
3. Navigate errors with arrow buttons or reload the page.

The overlay hooks global error events through the [`js` package](../api/js). No extra code is needed.

```go
func broken() {
    panic("boom")
}
```

@include\:ExampleFrame:{code:"/examples/components/runtime\_error\_component.go", uri:"/examples/runtime-error"}

**Notes**

* Only active in debug builds.
* Closing the overlay clears captured errors.
* Stores up to 15 errors per page.

## Development server

The `rfw dev` command launches a file-watching server that recompiles your project into WebAssembly on every change. It also serves generated assets and static files under `/`. If a `host/` directory is present, `rfw dev` builds and runs the host binary so you can test host components locally.

By default it listens on port `8080`. Override with `--port` or `RFW_PORT`.

```bash
rfw dev --port 8081 --debug
```

Flags:

* `--port` specify port
* `--host` expose to network
* `--debug` enable verbose logs and profiling endpoints

Environment variables:

* `RFW_PORT` set port
* `RFW_LOG_LEVEL` set log level (`debug`, `info`, `warn`, `error`)

Use `--host` to test on real devices. Avoid `--debug` in production—it serves profiling data.

## Hot reload

Changes to Go, RTML, Markdown and plugin assets trigger automatic rebuilds and refreshes. The watcher only recompiles what changed, keeping feedback loops short.

The browser keeps an `EventSource` connection to the development server. Whenever a rebuild completes for Go, RTML, Markdown, or plugin-managed assets, the server notifies connected tabs to reload automatically. Changes appear immediately without needing to refresh manually.

Some network file systems may miss events—restart the server or switch to polling in those cases.

## Profiling

With `--debug`, runtime metrics are exposed at `/debug/vars` and `/debug/pprof/`. The overlay adds two tabs:

* **Vars** shows `/debug/vars` as a searchable tree.
* **Pprof** lists profiles from `/debug/pprof/`.

Example:

```bash
go tool pprof http://localhost:8081/debug/pprof/profile
```

See the [Go pprof guide](https://pkg.go.dev/net/http/pprof) for more.

## Component tree

The overlay mirrors the active component hierarchy by capturing lifecycle events.

```js
// Dump the component tree
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_TREE()));

// Refresh manually
globalThis.RFW_DEVTOOLS_REFRESH();
```

**Notes**

* Active only in debug builds.
* Components added dynamically must be registered with [`AddDependency`](../api/core#dependency-injection) or use `rt-is`.

## Store inspection

Browse global stores under the **Store** tab.

```js
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_STORES()));
```

**Notes**

* Available only in debug builds.
* Only stores created during debug are visible.

## Signal inspection

The **Signals** tab lists reactive signals and their values.

```js
console.log(JSON.parse(globalThis.RFW_DEVTOOLS_SIGNALS()));
```

**Notes**

* Signals are anonymous and identified by ID.
* Available only in debug builds.

## Plugin inspection

The **Plugins** tab shows active build plugins and their configuration.

```js
console.log(globalThis.RFW_DEVTOOLS_PLUGINS());
```

**Notes**

* Only configured plugins appear.
* Available only in debug builds.

## Network inspector

The **Network** tab monitors HTTP requests made with [`http.FetchJSON`](../api/http).

```go
http.RegisterHTTPHook(func(start bool, url string, status int, d time.Duration) {
    if !start {
        log.Printf("%s -> %d in %v", url, status, d)
    }
})
```

@include\:ExampleFrame:{code:"/examples/components/fetchjson\_component.go", uri:"/examples/fetchjson"}

**Notes**

* Only tracks `FetchJSON` requests.
* Available in debug builds when `RFW_DEVTOOLS=1`.
