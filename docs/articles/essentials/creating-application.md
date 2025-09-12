# Creating an Application

This guide walks through the structure of a basic RFW application and explains how Go components, RTML templates, and the router come together to mount an interface in the browser.

## Project Layout

An RFW project pairs each Go component with an `.rtml` template and ships with a host folder for server components. The `rfw init` command scaffolds this layout for you:

```
hello-rfw/
├── main.go
├── components/
│   └── app_component.go
│   └── templates/
│       └── app_component.rtml
├── host/
│   ├── main.go
│   └── components/
│       └── hello_host.go
├── static/
│   └── (images, styles, etc.)
└── rfw.json
```

Files in `static/` are copied into `build/static/` and served from the root path. During development, requests to `/static/*` are mapped to `/*` so assets work without changing URLs.

## Bootstrapping

The entry point registers the root component with the router and starts it. Development mode enables helpful logging.

```go
package main

import (
    "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/router"
    "github.com/username/hello-rfw/components"
)

func main() {
    core.SetDevMode(true)
    router.RegisterRoute(router.Route{
        Path:      "/",
        Component: func() core.Component { return components.NewAppComponent() },
    })
    router.InitRouter()
    select {}
}
```

The generated `main.go` compiles to WebAssembly. To run it in the browser, include the Go runtime and instantiate the module with plain JavaScript—TypeScript builds are not supported yet:

```html
<script src="/wasm_exec.js"></script>
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load('/app.wasm', { go });
</script>
```

A future release will expose a global `rfw` object so that RFW APIs can be accessed directly from JavaScript without importing Go helpers.

## Defining the Root Component

Components embed `*core.HTMLComponent` and load their template at build time. The scaffolded root component looks like this:

```go
package components

import (
    _ "embed"

    "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/app_component.rtml
var appComponentTpl []byte

type AppComponent struct {
    *core.HTMLComponent
}

func NewAppComponent() *AppComponent {
    c := &AppComponent{
        HTMLComponent: core.NewHTMLComponent("AppComponent", appComponentTpl, nil),
    }
    c.SetComponent(c)
    c.AddHostComponent("HelloHost")
    c.Init(nil)
    return c
}
```

`NewHTMLComponent` wires the RTML template to the component, and `SetComponent` followed by `Init` prepares it for rendering. Once the router mounts the component, updates to its exported fields automatically patch the DOM.

Continue with the [Template Syntax](./template-syntax) guide to learn how RTML templates bind DOM output to Go state.
