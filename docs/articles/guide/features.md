# Why rfw?

**rfw** focuses on delivering a productive, Go-first workflow for the web. Instead of layering a virtual DOM over browser APIs, it updates the real DOM directly with **Selective DOM Patching**, a lightweight routine that mutates only what changed.

## Tag-Driven Composition

rfw v2 auto-wires your components from struct tags. No manual `Prop()` calls, no `AddDependency()`, no `SetOnMount()` boilerplate:

```go
type HomePage struct {
    composition.Component
    Count *t.Int  `rfw:"signal"`
    Name  *t.String `rfw:"inject"`
}

func (h *HomePage) Increment() { h.Count.Set(h.Count.Get() + 1) }
func (h *HomePage) OnMount()   { h.Count.Set(0) }
```

## Convention Over Configuration

Templates are found by struct name, `HomePage` → `HomePage.rtml`. Override with `rfw:"template:path"` when needed.

## Direct DOM Binding

Components render directly into DOM nodes. When state changes, rfw patches only the affected elements. No virtual tree is created, keeping the runtime small and predictable.

## Go-Centric Development

Both logic and templates live in Go modules:

* Reuse existing Go packages
* Benefit from static typing
* Use the standard Go toolchain

The `rfw` CLI handles scaffolding and WebAssembly builds.

## Reactive Signals & Stores

Fine-grained reactivity without the overhead:

* **Signals**: local component state via `rfw:"signal"` tags
* **Stores**: shared state across components via `rfw:"store:name"` tags
* **@expr:**: computed values inline in templates

## SSC (Server-Side Computed)

Render HTML on the server, hydrate in the browser. Faster time-to-content, better SEO, tighter security. Secrets stay on the server.

## Minimal Runtime

Only the parts you import are included. The JavaScript glue is tiny, most logic stays in Go, shipped as WebAssembly.

## Extensible Pipeline

Plugins extend the compiler and runtime for custom build steps, code generation, and browser API integrations.

## When to Use rfw

Use **rfw** when you want full control over output, prefer Go, or need to share code seamlessly between client and server.