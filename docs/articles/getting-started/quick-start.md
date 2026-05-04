# Quick Start

Walk through creating and running your first **rfw v2** application. Assumes your environment meets the [requirements](/docs/getting-started/requirements).

## Install the CLI

```bash
go install github.com/rfwlab/rfw/cmd/rfw@latest
rfw --version
```

## Scaffold a Project

```bash
rfw init github.com/username/hello-rfw
cd hello-rfw
```

Generated structure:

```
hello-rfw/
  main.go
  components/
  host/
  rfw.json
  go.mod
```

- `components/`, client-side components and `.rtml` templates
- `host/`, server-side host components (SSC)
- `rfw.json`, build config (SSC enabled by default)

Use `--skip-tidy` to skip `go mod tidy` during init.

## A First Component

In v2, `composition.New(&struct{})` auto-wires everything from struct tags. No manual template embedding, no `Prop()` calls, no `On()` registration.

### Counter.rtml

Templates are discovered by convention: the struct name maps to `StructName.rtml`. Place the template alongside your Go file or in any registered `embed.FS`.

```rtml
<root>
  <button @on:click:Increment>Count: {@expr:count}</button>
</root>
```

- `@on:click:Increment` fires the `Increment` method on the component
- `@expr:count` reads the `count` signal reactively

### Counter.go

```go
//go:build js && wasm

package main

import (
    "embed"

    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/router"
    "github.com/rfwlab/rfw/v2/types"
)

//go:embed Counter.rtml
var fs embed.FS

func init() {
    composition.RegisterFS(&fs)
}

type Counter struct {
    composition.Component

    Count *types.Int `rfw:"signal"`
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}

func (c *Counter) OnMount() {
    c.Count.Set(0)
}

func main() {
    router.Page("/", func() *types.View {
        return composition.New(&Counter{})
    })
    router.InitRouter()
    select {}
}
```

### What happened

| Tag | Effect |
|-----|--------|
| `rfw:"signal"` | Creates a `*types.Int` signal, registers it as prop `Count`, auto-wires reactivity |
| `Increment()` method | Auto-registered as DOM event handler, no `On()` call needed |
| `OnMount()` | Auto-discovered lifecycle hook, called when component mounts |
| `Counter.rtml` | Auto-found by struct name convention via `composition.RegisterFS` |
| `composition.Component` embed | Provides `Store()`, `History()`, and other composition helpers |

No `Template` field. No `WithTemplate()`. No `//go:embed` for template bytes. Tags do the work.

## Run the Development Server

```bash
rfw dev --debug
```

- Compiles Go to `app.wasm` (served as Brotli-compressed `app.wasm.br`)
- Serves static files under `/`
- Rebuilds and reloads on file changes
- Builds and runs host binary from `host/` when present

Flags:

- `--port`, set port (default 8080)
- `--host`, expose to network
- `--debug`, enable logs and profiling endpoints

Environment variables:

- `RFW_PORT`, set port
- `RFW_LOG_LEVEL`, set log level (`debug`, `info`, `warn`, `error`)

## Build for Production

```bash
rfw build
```

Outputs:

- `build/client/`, Wasm bundle and assets
- `build/static/`, copied static files
- `build/host/`, host binary for SSC

Production builds use `-trimpath` and `-ldflags="-s -w"` to strip debug info. Export `RFW_SKIP_STRIP=1` to keep symbols. During development, `rfw dev` sets `RFW_DEV_BUILD=1`, enabling the `rfwdev` build tag.

## What You Learned

- Installing the CLI and scaffolding a project
- Creating a component with tag-driven composition
- Templates found by convention (`StructName.rtml`)
- Registering routes with `router.Page()`
- Running the dev server and building for production

## Next Steps

- [Router](/docs/guide/router)
- [Signals and Stores](/docs/guide/store-vs-signals)
- [SSC](/docs/guide/ssc)