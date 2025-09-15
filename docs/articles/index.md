# Introduction

Welcome to the documentation for **rfw v1**. Use the search box in the top navigation to quickly find what you need.

## What is rfw?

rfw (Reactive Framework for the Web) is a framework for building user interfaces with **Go** and **WebAssembly**. It extends standard HTML, CSS, and JavaScript with a declarative, component-based model that lets you write your UI entirely in Go. State in Go is bound directly to the DOM: when values change, rfw updates the page automatically—without a virtual DOM.

Here is the simplest example using the **Composition API**:

```go
package main

import (
    _ "embed"
    "context"

    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/composition"
    "github.com/rfwlab/rfw/v1/state"
)

//go:embed counter.rtml
var tpl []byte

func NewCounter() *core.HTMLComponent {
    cmp := composition.Wrap(core.NewComponent("Counter", tpl, nil))

    count := state.NewSignal(0)
    cmp.Prop("count", count)

    cmp.On("increment", func() {
        count.Set(count.Get() + 1)
    })

    return cmp.HTML()
}
```

```rtml
<root>
  <button @on:click:increment>Count is: {count}</button>
</root>
```

**Result:** clicking the button increases the counter by updating a reactive signal bound to the template.

This example shows the two core features of rfw:

* **Declarative rendering**: RTML templates extend HTML with state placeholders and event bindings.
* **Reactivity**: signals track state changes and trigger DOM updates automatically.

## A Flexible Framework

rfw adapts to the way you want to build:

* **Progressive enhancement** – drop a component into any HTML page with no setup.
* **Full applications** – build SPAs with routing, state, and live reload.
* **Hybrid rendering** – pre-render on the server, hydrate in the browser.
* **Advanced targets** – integrate with WebGL, workers, or browser APIs when needed.

You can start with a single reactive widget and grow into a full application—without switching frameworks.

## RTML Templates

Components use **RTML** (Reactive Template Markup Language). The `.rtml` file holds markup, the Go file holds logic. At build time, templates are embedded in your binary and load instantly in the browser.

Learn more about [Template Syntax](/docs/essentials/template-syntax).

## Component Styles

rfw supports two styles:

### Composition API

Define state with signals and actions inside a setup-like function, using helpers from the `composition` package. This is the recommended style for new code.

### Struct Components

Embed `*core.HTMLComponent` in a struct to manage state via exported fields and methods:

```go
type Counter struct {
    *core.HTMLComponent
    Count int
}

func New() *Counter {
    c := &Counter{}
    c.HTMLComponent = core.NewComponentWith("Counter", tpl, nil, c)
    return c
}

func (c *Counter) Increment() { c.Count++ }
```

### Functional Components

For very small pieces of UI, return a component directly:

```go
func NewCounter() *core.HTMLComponent {
    return core.NewComponent("Counter", tpl, nil)
}
```

### Which to Choose?

* **Composition API**: recommended for most projects; flexible, expressive, and scales well.
* **Struct components**: useful if you prefer method receivers and lifecycle hooks.
* **Functional components**: best for simple presentational elements.

## Next Steps

* [Quick Start](./getting-started/quick-start) – build your first component in minutes.
* [The Guide](./guide/features) – explore rfw in depth.
* [The Essentials](./guide/creating-application) – learn the essential basis of rfw.
* [API Reference](../api/core) – see how everything works under the hood.
* [Contributing](../../CONTRIBUTING.md) – help improve the framework.
