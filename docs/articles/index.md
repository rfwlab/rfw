# Introduction

You are reading the documentation for rfw v1.

## What is rfw?

rfw (Reactive Framework for the Web) is a framework for building user interfaces with Go and WebAssembly. It builds on standard HTML, CSS, and JavaScript and provides a declarative, component-based programming model that lets you author your UI entirely in Go. rfw binds Go data structures directly to DOM nodes, so when state changes the framework updates the page without relying on a virtual DOM.

Here is a minimal example:

```go
package main

import (
    _ "embed"

    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/dom"
    "github.com/rfwlab/rfw/v1/state"
)

//go:embed counter.rtml
var tpl []byte

func NewCounter() *core.HTMLComponent {
    store := state.NewStore("counter")
    store.Set("count", 0)
    dom.RegisterHandlerFunc("increment", func() {
        current, _ := store.Get("count").(int)
        store.Set("count", current+1)
    })
    c := core.NewComponent("Counter", tpl, nil)
    c.Init(store)
    return c
}
```

```html
<root>
  <button @click:increment>Count is: {count}</button>
</root>
```

**Result**

Mounting this component displays a button whose label increments with each click through a registered handler that updates the store directly.

The example highlights two core features of rfw:

- **Declarative rendering**: RTML templates extend HTML with placeholders for state and tokens for events.
- **Reactivity**: rfw tracks store mutations and selectively patches the DOM nodes that changed.

## The Modern Framework

rfw does not force you into a single style of development. You can start small and grow as needed, mixing approaches without rewriting your project. Depending on what you build, rfw can adapt to different roles:

- **Progressive enhancement** – drop a component into an existing HTML page and it just works, no complex setup.
- **Full applications** – build full SPAs (single-page applications) with routing, state management and live reloading during development.
- **Hybrid rendering** – pre-render pages on the server with the host utilities, then let components hydrate and become interactive in the browser.
- **Advanced targets** – hook into WebGL, web workers, or specialized browser APIs when you need raw performance or parallelism.

rfw is designed to scale from a single reactive widget to entire applications, letting you choose the right level of complexity for your project.

## RTML Templates

Most projects author components using **RTML** (Reactive Template Markup Language). An RTML file encapsulates the component's markup while the Go file holds its logic. During builds the CLI embeds templates into your binaries so they load instantly in the browser.

## Component Styles

rfw components can be written in two different styles.

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

For small pieces of UI, return `*core.HTMLComponent` directly from a function:

```go
func NewCounter() *core.HTMLComponent {
    c := core.NewComponent("Counter", tpl, nil)
    return c
}
```

### Which to Choose?

- Use **struct components** when you need lifecycle hooks, internal state or methods.
- Use **functional components** for simple presentational elements.

## Still Got Questions?

Read the [architecture overview](./architecture) or open an issue on GitHub if something isn't clear.

## Pick Your Learning Path

- [Try the Quick Start](./getting-started/quick-start) for a hands-on introduction.
- [Read the Guide](./guide/features) to explore the framework in depth.
- [Browse the API](../api/core) to see how rfw works under the hood.
