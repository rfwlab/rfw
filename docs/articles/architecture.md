# Architecture

RFW's goal is to let developers build reactive web interfaces in pure Go. The runtime runs Go in the browser via WebAssembly while templates are written in **RTML**, an HTML-like language that binds directly to Go state. The framework wires data, events and rendering into a single reactive loop so that UI updates automatically follow changes in Go variables.

By compiling to WebAssembly, the same Go code can run on both the client and the server, enabling shared types and business logic. Components are small, reusable units that encapsulate state and behaviour; the router simply mounts them without extra glue code.

## Minimal counter example

```rtml
<root>
  <button @click:increment>Count: {count}</button>
</root>
```

```go
package counter

import (
    "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/dom"
)

//go:embed counter.rtml
var tpl []byte

func New() *core.HTMLComponent {
    c := core.NewComponent("Counter", tpl, map[string]any{"count": 0})
    dom.RegisterHandlerFunc("increment", func() {
        c.Props["count"] = c.Props["count"].(int) + 1
    })
    return c
}
```

The `@click` directive binds the `increment` handler to the button. When the button is pressed, the Go value is increased and the DOM is patched automatically—no manual manipulation is required. RTML placeholders such as `{count}` are resolved against the component's `Props` map on every re-render.

For larger components consider updating state through stores to avoid race conditions. The Go WASM runtime is single threaded; heavy computations should be offloaded to web workers or handled asynchronously to prevent UI freezing.

## Prerequisites

You should be comfortable with basic HTML, CSS, JavaScript and Go before using RFW. Familiarity with concepts such as reactivity and the Go module system helps in structuring larger applications.

Because Go's WebAssembly runtime does not support preemptive multitasking, avoid long blocking calls in handlers. JavaScript integration currently uses plain JavaScript; TypeScript is not yet supported, so wrappers may be required in mixed projects.
