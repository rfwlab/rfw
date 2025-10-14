# Quick Start

This guide walks through creating and running your first **rfw** application. It assumes your environment meets the [requirements](/docs/getting-started/requirements).

## Install the CLI

```bash
go install github.com/rfwlab/rfw/cmd/rfw@latest
rfw --version
```

The CLI builds, runs, and serves your project.

## Scaffold a Project

Initialize a new project:

```bash
rfw init github.com/username/hello-rfw
cd hello-rfw
```

The generated structure includes:

```
hello-rfw/
  main.go
  components/
  host/
  rfw.json
  go.mod
```

* `components/` holds client-side components and `.rtml` templates.
* `host/` holds optional server-side (host) components.
* `rfw.json` configures build options such as SSC.

Use `--skip-tidy` to skip `go mod tidy` during init.

## A First Component

### counter.rtml

```rtml
<root>
  <button @on:click:increment>Count is: {count}</button>
</root>
```

### counter.go

```go
package main

import (
    "github.com/rfwlab/rfw/v1/composition"
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/router"
    "github.com/rfwlab/rfw/v1/state"
)

//go:embed counter.rtml
var tpl []byte

func NewCounter() *core.HTMLComponent {
    cmp := composition.Wrap(core.NewComponent("Counter", tpl, nil))
    count := state.NewSignal(0)

    cmp.Prop("count", count)
    cmp.On("increment", func() { count.Set(count.Get() + 1) })

    return cmp.HTML()
}

func main() {
    router.RegisterRoute(router.Route{
        Path: "/",
        Component: func() core.Component { return NewCounter() },
    })
    router.InitRouter()
    select {}
}
```

Open the app in a browser and click the button—the counter increments on each click.

## Run the Development Server

```bash
rfw dev --debug
```

Features:

* Compiles Go sources to `app.wasm` (served as the Brotli-compressed `app.wasm.br` bundle)
* Serves static files under `/`
* Rebuilds and reloads on file changes
* Runs host components from `host/` if present

Flags:

* `--port` set port (default 8080)
* `--host` expose to network
* `--debug` enable logs and profiling endpoints

Environment variables:

* `RFW_PORT` set port
* `RFW_LOG_LEVEL` set log level (`debug`, `info`, `warn`, `error`)

## Build for Production

```bash
rfw build
```

The build command outputs:

* `build/client/` – Wasm bundle and assets
* `build/static/` – copied static files
* `build/host/` – companion host binary

Outside of debug mode the build uses Go's `-trimpath` and `-ldflags="-s -w"` flags to strip debug information from the generated WebAssembly module.

Need to inspect symbols instead? Export `RFW_SKIP_STRIP=1` so `go build` keeps debug metadata intact. During development `rfw dev` automatically sets `RFW_DEV_BUILD=1`, enabling the `rfwdev` build tag and development helpers without affecting production builds.

## What You Learned

* Installing the CLI
* Scaffolding a project
* Creating and mounting a component
* Running the dev server with hot reload
* Building for production

## Next Steps

* [Templates](/docs/essentials/template-syntax)
* [Signals and Stores](/docs/essentials/signals-and-effects)
* [Components](/docs/essentials/components-basics)
* [Architecture](/docs/architecture)
