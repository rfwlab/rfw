<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/rfwlab/brandbook/refs/heads/main/logos/full/png/light-full.png">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/rfwlab/brandbook/refs/heads/main/logos/full/png/dark-full.png">
  <img alt="Logo" src="https://raw.githubusercontent.com/rfwlab/brandbook/refs/heads/main/logos/full/png/dark-full.png" height="100">
</picture>

<hr />

### Real-time dashboards and internal tools, written entirely in Go. No JavaScript. No glue code.

![rfw counter demo with live updates](docs/assets/hero-counter.gif)

[Documentation](./docs/articles/index.md)
</div>

rfw is "Phoenix LiveView for Go". It lets you build interactive, real-time web apps using Server Side Computed (SSC) components. 

Instead of writing a REST API and a frontend framework, you write Go. rfw handles the WebSocket synchronization and DOM updates for you. It is ideal for:
- Real-time dashboards
- Internal admin tools
- Control planes
- Any app where server state needs to reflect instantly in the UI

## Why rfw?

If you are using `templ` + `htmx` (or `datastar`), you are already moving toward server-driven UI. rfw takes this further by providing a full state-synchronization engine. You get the productivity of a frontend framework (like React or Vue) but with the simplicity of a single Go binary and type-safe end to end.

## Getting Started

```bash
go install github.com/rfwlab/rfw/cmd/rfw@latest
rfw init github.com/user/app
rfw dev
```

Minimal hello-world component:

```go
package main

import (
    "embed"

    "github.com/rfwlab/rfw/v1/composition"
    cmp "github.com/rfwlab/rfw/v1/components"
)

//go:embed templates/hello.rtml
var helloTpl []byte

func main() {
    composition.Wrap(cmp.New("Hello", helloTpl))
}
```

`templates/hello.rtml`:

```html
<h1>Hello {{ .Name }}</h1>
```

By default the development server listens on port `8080`. Override it with
the `--port` flag or the `RFW_PORT` environment variable:

```bash
RFW_PORT=3000 rfw dev
```

Control server log verbosity with the `RFW_LOG_LEVEL` environment variable.
Possible values are `debug`, `info`, `warn`, and `error` (default is `info`):

```bash
RFW_LOG_LEVEL=debug rfw dev
```

Enable the in-browser debugging overlay with the `--debug` flag:

```bash
rfw dev --debug
```

Use `Ctrl`+`Shift`+`D` in the browser to toggle the overlay that shows the
component tree and console logs with basic runtime metrics.

## Server Side Computed (SSC)

SSC is the core of rfw. Most application logic runs on the server, while the browser loads a lightweight binary to hydrate the HTML. The server and client keep state synchronized through a persistent WebSocket connection. 

Components use host signal types (`t.HInt`, `t.HString`, etc.) to declare server-synced bindings. See the [SSC guide](./docs/articles/guide/ssc.md) for more details.

## Testing

Run all tests with:

```bash
go test ./...
```

Continuous Integration runs the same command on every push. See the [testing guide](./docs/articles/testing.md) for more details.


## Build-level Plugins

`rfw` exposes a simple plugin system for build-time tasks. The CLI
automatically detects `PreBuild`, `Build` and `PostBuild` methods on plugins
and invokes them when present. Each plugin must still provide a file-watcher
trigger via `ShouldRebuild` and a numeric `Priority` to determine execution
order.

### Tailwind CSS

`rfw` includes a build step for [Tailwind CSS](https://tailwindcss.com/) using the official standalone CLI.
Place an `input.css` file (commonly under `static/`) containing the `@tailwind` directives in your project. During development the server watches
template, stylesheet and configuration files and emits a trimmed `tailwind.css`
containing only the classes you use, without requiring Node or a CDN.

### File-based Routing

The built-in `pages` plugin scans a `pages/` directory and automatically
registers routes based on its structure. Each Go file maps to a URL path:

```
pages/
  index.go        // -> /
  about.go        // -> /about
  posts/[id].go   // -> /posts/:id
```

Every file must expose a constructor using the PascalCase form of its path,
such as `func About() core.Component`. The plugin generates a temporary
`routes_gen.go` that calls `router.RegisterRoute` for each page during the
build. Import the generated package to execute the registrations, typically
via a blank import in your entrypoint:

```
import _ "your/module/pages"
```

A working example can be found under `docs/examples/pages`, which
contains `index.go`, `about.go` and `posts/id.go` demonstrating the
generated routes `/`, `/about` and `/posts/:id`. The documentation site in
this repository uses the `pages` plugin for its home (`/`) and about (`/about`)
pages, while the `docs` plugin continues to power the documentation
content itself.

For more details and best practices, see the [Pages Plugin guide](./docs/articles/plugins/pages.md).

---
*rfw uses WebAssembly (Wasm) to bridge the server-client gap, but you only ever write Go.*
