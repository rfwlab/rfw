# The mental model

rfw is one idea applied consistently: **the DOM is a projection of Go
state**. Everything else in the framework is plumbing for that idea.

- A component is a Go struct wrapping `*core.HTMLComponent`, paired with an
  `.rtml` template embedded in the binary.
- State lives in Go: stores (`state.NewStore`) and signals
  (`state.NewSignal`). Templates bind to them with directives such as
  `@store:module.store.key`, `@signal:name`, `@expr:`, `@if:`/`@endif` and
  `@for:`/`@endfor`.
- Rendering substitutes those directives with the current values and marks
  the elements; when a store key or signal changes, only the marked
  elements update. You never touch the DOM to reflect state.
- Events go the other way: `@on:click:increment` in the template becomes a
  delegated `data-on-click` attribute resolved to a handler you registered
  from Go. No listeners per element, no JavaScript.
- The whole program, components, state, handlers, router, compiles to a
  single wasm binary that runs in the browser.

## A complete component

`templates/counter.rtml`:

```html
<root>
  <p>Clicked @store:app.counter.count times</p>
  <button @on:click:increment>+1</button>
</root>
```

`counter.go`:

```go
//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

//go:embed templates/counter.rtml
var counterTpl []byte

var counter = func() *state.Store {
	s := state.NewStore("counter", state.WithModule("app"))
	s.Set("count", 0)
	return s
}()

func NewCounter() *core.HTMLComponent {
	c := core.NewHTMLComponent("Counter", counterTpl, nil)
	c.SetComponent(c)
	c.Init(nil)

	dom.RegisterHandlerFunc("increment", func() {
		n, _ := counter.Get("count").(int)
		counter.Set("count", n+1)
	})
	return c
}
```

Trace the loop: the button click bubbles to the component root, the
delegated `data-on-click` resolves to the `increment` handler, the handler
mutates the store, the store notifies the `@store` binding, the `<span>`
holding the count updates. State moved; the DOM followed.

## If you come from Node

There is no build pipeline to assemble: no npm, no bundler, no transpiler,
no `node_modules`; `rfw dev` compiles your Go module and serves it, and the
deliverable is one wasm binary plus static files. There is also no
client/server language split to manage: when you add Server Side Computed
components, both halves are Go sharing the same types, so "the API contract"
is just a function signature the compiler checks.
