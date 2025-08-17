# Getting Started

rfw ships with a small CLI that scaffolds new projects and drives the
build system. Install it and spin up a sample application:

```bash
go install github.com/rfwlab/rfw-cli@latest
rfw-cli init github.com/username/hello-rfw
cd hello-rfw
rfw-cli dev
```

The `dev` command compiles your Go code to WebAssembly and launches a
development server with live reloading.

## A tiny demo

Templates in rfw are written in **RTML** – an HTML-like language with
reactive bindings. Create `counter.rtml`:

```rtml
<root>
  <button on:click="count++">Count: {count}</button>
</root>
```

And the accompanying component:

```go
package counter

import "github.com/rfwlab/rfw/v1/core"

//go:embed counter.rtml
var tpl []byte

type Counter struct {
    *core.HTMLComponent
    Count int
}

func New() *Counter {
    c := &Counter{HTMLComponent: core.NewHTMLComponent("Counter", tpl, nil)}
    c.SetComponent(c)
    c.Init(nil)
    return c
}
```

Running `rfw-cli dev` now gives you a working counter with almost no
JavaScript written by hand.
