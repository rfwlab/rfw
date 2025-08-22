# Why rfw?

rfw focuses on delivering a productive Go‑first workflow for the web.
Instead of layering a virtual DOM over browser APIs, the framework maps
state changes directly to the real DOM. A lightweight routine named
**Selective DOM Patching** compares new markup with existing nodes and
mutates them in place. Below are some of the ideas that make it stand out.

## Direct DOM binding

Every component renders straight into real DOM nodes. When a piece of
state changes rfw applies Selective DOM Patching to update only the
affected elements. No separate virtual tree is created, keeping the
runtime small, predictable and easy to reason about.

## Go‑centric development

Both application logic and templates live in Go modules. You can reuse
existing packages, benefit from static types and leverage the standard
toolchain. The `rfw` command wraps common tasks like project
scaffolding and building WebAssembly binaries.

## Reactive stores

State is organised in named stores. Components subscribe to keys and are
automatically re‑rendered when values change. Stores support **computed
values** derived from other keys and **watchers** that react to
mutations, enabling complex data flows with minimal boilerplate.

## Minimal runtime

Only the framework pieces you import end up in the bundle. The generated
JavaScript glue code is tiny and the majority of your logic remains in
Go, shipped as WebAssembly.

## Extensible pipeline

Plugins can augment both the compiler and runtime. They allow custom
build steps, additional code generation or integrations with browser
APIs. The plugin system keeps the core lightweight while still enabling
advanced use cases.

## When to use rfw

rfw shines when you want tight control over the generated output, prefer
writing in Go or need to share code between client and server. While the
project is experimental, it demonstrates how a simple reactive model can
produce interactive interfaces without a large JavaScript framework.
The main component highlights several core features.

@include:ExampleFrame:{code:"/examples/components/main_component.go", uri:"/examples/main"}
