# Components Basics

Components are the building blocks of every RFW application. A component pairs a Go struct with an RTML template and exposes reactive state through exported fields. When those fields change, the DOM updates automatically.

## Defining a Component

Create a struct that embeds `*core.HTMLComponent` and register its template. The `SetComponent` call wires the struct so RFW can track exported fields.

```go
package components

import (
  _ "embed"

  "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/counter.rtml
var counterTpl []byte

type Counter struct {
  *core.HTMLComponent
  Count int
}

func NewCounter() *Counter {
  c := &Counter{HTMLComponent: core.NewHTMLComponent("Counter", counterTpl, nil)}
  c.SetComponent(c)
  c.Init(nil)
  return c
}
```

The matching `counter.rtml` template references `Count` directly and responds when it changes.

```rtml
<div>
  <button @on:click:inc>-</button>
  <span>{Count}</span>
  <button @on:click:dec>+</button>
</div>
```

## Props

Child components receive data through the `Props` map passed to `NewHTMLComponent` or via `@include:Child:{prop:"value"}` in a template. Inside the component, call `GetProp("prop")` to read it. Props are immutable; to communicate back to the parent use events or shared stores.

## Methods and Events

Any exported method may be invoked from a template using the `@on:` directive. In the counter example, handlers can update `Count` directly:

```go
func (c *Counter) Inc() { c.Count++ }
func (c *Counter) Dec() { c.Count-- }
```

Exposing only the methods you want templates to call keeps encapsulation clear.

## Cleanup on Unmount

When a component is removed from the DOM, `OnUnmount` lets you release resources like watchers or timers. `Store.RegisterWatcher` returns a cleanup function that should be called in this hook:

```go
type Counter struct {
  *core.HTMLComponent
  stop func()
}

func (c *Counter) OnMount() {
  w := state.NewWatcher([]string{"Count"}, func(m map[string]any) {
    log.Println("count changed", m["Count"])
  })
  c.stop = c.Store.RegisterWatcher(w)
}

func (c *Counter) OnUnmount() {
  c.stop()
}
```

`OnUnmount` runs before the component is detached, ensuring cleanup occurs while the component still exists.

## Composing Components

Components can nest by including each other in RTML. Slots allow parents to inject markup into predefined outlets of a child component. This enables flexible layouts while keeping logic isolated.

Understanding components is key to structuring RFW apps. The following chapters build on this foundation to explore reactivity and data flow in more depth.
