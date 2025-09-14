# Introduction

You are reading the documentation for rfw.

## What is rfw?

rfw (Reactive Framework for the Web) is a progressive framework for building user interfaces with Go and WebAssembly. It builds on top of standard HTML, CSS, and JavaScript and provides a declarative, component-based programming model that lets you author your entire UI in Go. When compiled to Wasm, rfw runs in the browser with minimal runtime overhead.

Templates are written in **RTML**, an HTML-like language extended with directives that bind the DOM to Go data or events. When state changes, rfw automatically updates the affected DOM nodes—no manual DOM manipulation is needed.

## A Minimal Example

Here is a minimal counter component:

```rtml
<root>
  <button @on:click:increment>
    Count is: @signal:count
  </button>
</root>
```

```go
package counter

import (
    "github.com/rfwlab/rfw/v1/composition"
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/router"
    "github.com/rfwlab/rfw/v1/state"
)

//go:embed counter.rtml
var tpl []byte

type Counter struct {
    *composition.Component
    count *state.Signal[int]
}

func New() *Counter {
    cmp := composition.Wrap(core.NewComponent("Counter", tpl, nil))
    c := &Counter{Component: cmp, count: state.NewSignal(0)}
    cmp.Prop("count", c.count)
    cmp.On("increment", func() { c.count.Set(c.count.Get() + 1) })
    return c
}

func main() {
    router.RegisterRoute(router.Route{
        Path: "/",
        Component: func() core.Component { return New() },
    })
    router.InitRouter()
    select {}
}
```

Mounting this component in the browser shows a button whose label increments each time it is clicked. The example demonstrates two core features:

- **Declarative Rendering** – the RTML template describes the DOM based on Go state.
- **Reactivity** – rfw tracks mutations and efficiently patches only the nodes that changed.

You may already have questions—don't worry. The rest of the documentation explores these concepts in detail.

## Prerequisites

The guides assume basic familiarity with:

- HTML and CSS
- JavaScript
- Go

Browser integration currently uses plain JavaScript; TypeScript isn't supported yet. A future release will expose a global `rfw` object to let you call rfw APIs directly from JavaScript.

## The Progressive Framework

The web spans a wide range of use cases. rfw is designed to be flexible and adoptable incrementally:

- **Enhance existing pages** by mounting individual components into server-rendered HTML.
- **Build single-page applications** using the router and dev server.
- **Render on the server** with Host Components and hydrate on the client.
- **Target other environments** such as desktop or mobile shells that can run WebAssembly.

Whichever entry point you choose, the core knowledge of components, reactivity, and templates remains the same. As your application grows, the skills you learn continue to apply—rfw grows with you.

## Single-File Components

Most rfw projects pair a Go source file with an `.rtml` template:

```
counter.go
counter.rtml
```

During a build, the `rfw` CLI embeds the template into the component and outputs an optimized Wasm bundle. This single-file style keeps a component's logic, markup, and styles close together while leveraging Go's tooling and type safety.

## Next Steps

Continue with the [Quick Start](./quick-start) to scaffold your first project and learn how to compile and serve the Wasm bundle. The sidebar lists additional guides covering templates, state management, components, and more.
