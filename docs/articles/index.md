# Introduction

Welcome to the documentation for **rfw v2**. Use the search box in the top navigation to quickly find what you need.

## What is rfw?

rfw (Reactive Framework for the Web) is a framework for building user interfaces with **Go** and **WebAssembly**. It extends standard HTML with a declarative, component-based model that lets you write your UI entirely in Go. State in Go is bound directly to the DOM: when values change, rfw updates the page automatically—without a virtual DOM.

Here is the simplest example:

```go
//go:build js && wasm

package main

import (
    "embed"

    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

//go:embed templates
var templates embed.FS

func init() {
    composition.RegisterFS(&templates)
}

type Counter struct {
    composition.Component
    Count t.Int
}

func (c *Counter) Increment() { c.Count.Set(c.Count.Get() + 1) }

func NewCounter() *t.View {
    view, _ := composition.New(&Counter{Count: *t.NewInt(0)})
    return view
}
```

```rtml
<root>
  <button @on:click:Increment>Count: @signal:Count</button>
  <p>@expr:Count.Get * 2 doubled</p>
</root>
```

**Result:** clicking the button increases the counter. The `@expr:` directive computes derived values inline.

This example shows the core features of rfw v2:

* **Type-based composition**, field types (`t.Int`, `*t.Store`, `*t.Ref`, etc.) auto-wire reactivity
* **Convention over configuration**, templates found by struct name (`Counter` → `Counter.rtml`) or `Template()` method
* **Auto-discovered lifecycle**, `OnMount()`/`OnUnmount()` methods detected automatically
* **Declarative RTML**, `@signal:`, `@on:click:`, `@expr:` bind state and events

## Key Concepts

* **Signals**, reactive values (`t.Int`, `t.String`, etc.) that update the DOM when changed
* **Stores**, shared state across components, scoped by module and name
* **Composition**, `composition.New(&struct{})` auto-wires everything from field types
* **Routing**, `router.Page()` and `router.Group()` for client-side navigation
* **SSC**, Server-Side Computed: the host renders HTML, Wasm hydrates it

## Next Steps

* [Quick Start](./getting-started/quick-start), build your first component in minutes
* [Creating an Application](./essentials/creating-application), project structure and bootstrap
* [Components Basics](./essentials/components-basics), components, types, and lifecycle
* [Template Syntax](./essentials/template-syntax), RTML directives and expressions
* [API Reference](./api/composition), full API docs

## Getting Started

* [Quick Start](./getting-started/quick-start)
* [Requirements](./getting-started/requirements)

## Essentials

* [Creating an Application](./essentials/creating-application)
* [Components Basics](./essentials/components-basics)
* [Composition](./essentials/composition)
* [Template Syntax](./essentials/template-syntax)
* [Reactivity Fundamentals](./essentials/reactivity-fundamentals)
* [Signals, Effects & Watchers](./essentials/signals-effects-and-watchers)
* [Event Handling](./essentials/event-handling)
* [Lifecycle Hooks](./essentials/lifecycle-hooks)
* [Conditional Rendering](./essentials/conditional-rendering)
* [List Rendering](./essentials/list-rendering)
* [Computed Properties](./essentials/computed-properties)
* [Template Refs](./essentials/template-refs)
* [Class and Style Bindings](./essentials/class-and-style-bindings)

## Guides

* [Router](./guide/router)
* [SSC (Server-Side Computed)](./guide/ssc)
* [State Management](./guide/state-management)
* [Store vs Signals](./guide/store-vs-signals)
* [Provide & Inject](./guide/provide-inject)
* [Features](./guide/features)
* [Manifest](./guide/manifest)
* [CLI](./guide/cli)

## API Reference

* [composition](./api/composition)
* [router](./api/router)
* [state](./api/state)
* [signal](./api/signal)
* [rtml](./api/rtml)
* [core](./api/core)
* [dom](./api/dom)
* [events](./api/events)

## Migration

* [From Vue](./migration/from-vue)
* [From React](./migration/from-react)
* [From Svelte](./migration/from-svelte)
* [Comparison](./migration/comparison)