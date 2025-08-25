# Architecture

RFW's goal is to let developers build reactive web interfaces in pure Go. The runtime runs Go in the browser via WebAssembly while templates are written in **RTML**, an HTML-like language that binds directly to Go state.

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

Clicking the button increments the Go value and the framework updates the text automatically—no manual DOM manipulation is required.

## Prerequisites

You should be comfortable with basic HTML, CSS, JavaScript and Go before using RFW. JavaScript integration currently uses plain JavaScript; TypeScript is not yet supported.

