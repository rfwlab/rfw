# Architecture

rfw lets you build reactive web interfaces entirely in **Go**. The runtime executes Go in the browser via WebAssembly, while templates are written in **RTML**, an HTML-like language that binds directly to Go state. Data, events, and rendering flow together in a single reactive loop—UI updates follow automatically when variables change.

By compiling to WebAssembly, the same Go code can run on both client and server, enabling shared types and business logic. Components are small reusable units that encapsulate state and behaviour, mounted by the router without extra glue code.

## Minimal counter example

```rtml
<root>
  <button @on:click:increment>Count: {count}</button>
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

Clicking the button calls the `increment` handler. The Go value updates and the DOM is patched—no manual manipulation required. RTML placeholders like `{count}` resolve against the component’s props on every render.

For larger components, prefer **stores** to manage state and avoid race conditions. Since Go’s WASM runtime is single-threaded, offload heavy work to web workers or async tasks to keep the UI responsive.

## Prerequisites

Before using rfw, you should know basic HTML, CSS, JavaScript, and Go. Familiarity with reactivity and the Go module system helps when structuring bigger projects.

Notes:

* Avoid long blocking calls in handlers—Go’s WebAssembly runtime lacks preemptive multitasking.
* JavaScript integration uses plain JS; TypeScript is not yet supported.
