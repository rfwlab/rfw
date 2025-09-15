# Creating an Application

This guide explains the structure of an **rfw** project and how Go components, RTML templates, and the router combine to render your application in the browser.

---

## Project Layout

Running `rfw init` scaffolds a project like this:

```
hello-rfw/
├── main.go
├── components/
│   ├── app_component.go
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

* `components/` contains Go components and their `.rtml` templates.
* `host/` holds optional server‑side components.
* `static/` files are copied into `build/static/` and served at `/`. During `rfw dev`, `/static/*` resolves to `/*` so you don’t need to adjust URLs.
* `rfw.json` defines build configuration.

---

## Bootstrapping

The entry point registers a root component with the router and starts it. In development mode, logging and debug helpers are enabled.

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
        Path: "/",
        Component: func() core.Component { return components.NewAppComponent() },
    })
    router.InitRouter()
    select {}
}
```

The compiled `main.go` becomes a WebAssembly bundle. To run it, load the Go runtime and the generated module:

```html
<script src="/wasm_exec.js"></script>
<script src="/wasm_loader.js"></script>
<script>
  const go = new Go();
  WasmLoader.load('/app.wasm', { go });
</script>
```

> TypeScript support is not yet available. A future release will expose a global `rfw` object to call APIs directly from JavaScript.

---

## Defining the Root Component

Components embed `*core.HTMLComponent` and bind to a template. The scaffolded root component looks like this:

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

* `NewHTMLComponent` links the template to the component.
* `SetComponent` + `Init` prepare it for rendering.
* `AddHostComponent` registers a server‑side component that can interact with the client.

When mounted, changes to exported fields trigger reactive DOM updates automatically.

---

## Next Steps

Continue with [Template Syntax](./template-syntax) to learn how RTML templates bind DOM output to Go state.
