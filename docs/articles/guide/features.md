# Why rfw?

**rfw** focuses on delivering a productive, Go-first workflow for the web. Instead of layering a virtual DOM over browser APIs, it updates the real DOM directly with **Selective DOM Patching**, a lightweight routine that mutates only what changed.

## Direct DOM Binding

Components render directly into DOM nodes. When state changes, rfw patches only the affected elements. No virtual tree is created, keeping the runtime small and predictable.

## Go-centric Development

Both logic and templates live in Go modules:

* Reuse existing Go packages
* Benefit from static typing
* Use the standard Go toolchain

The `rfw` CLI handles scaffolding and WebAssembly builds.

## Reactive Stores

State lives in named stores. Components subscribe to keys and re-render automatically:

* **Computed values**: derive data from other keys
* **Watchers**: trigger side effects on mutation

This supports complex flows with minimal boilerplate.

## Minimal Runtime

Only the parts you import are included. The JavaScript glue is tiny—most logic stays in Go, shipped as WebAssembly.

## Extensible Pipeline

Plugins extend the compiler and runtime for:

* Custom build steps
* Code generation
* Browser API integrations

This keeps the core lean but flexible.

## When to Use rfw

Use **rfw** when you want full control over output, prefer Go, or need to share code seamlessly between client and server. It demonstrates how a simple reactive model can power interactive UIs without relying on a heavy JavaScript framework.

@include\:ExampleFrame:{code:"/examples/components/main\_component.go", uri:"/examples/main"}
