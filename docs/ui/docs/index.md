# Introduction

rfw is a reactive framework written in Go and compiled to WebAssembly.
It focuses on a direct binding model between state and the DOM, avoiding
virtual DOM diffing and embracing an event‑driven architecture. The
project is experimental and aims to provide a pleasant way to build web
interfaces with Go.

## What is rfw?

At its core rfw glues Go data structures directly to DOM nodes.
Components describe their markup in **RTML**, an HTML‑like template
language, while state lives in reactive stores. When a value in a store
changes the framework updates every subscribed component automatically.
No intermediate JavaScript is required and the browser only executes the
minimal runtime that wires things together.

## Core ideas

- **Direct DOM bindings** – state updates are written straight to the
  affected nodes without virtual DOM diffing.
- **Reactive stores** – computed values and watchers keep application
  state consistent with very little boilerplate.
- **Go‑first tooling** – projects use the standard Go toolchain and the
  `rfw` utility for compilation and development tasks.

## Next steps

If you are new to the framework start with the
[Getting Started guide](./getting-started.md) which walks through the
CLI and the creation of a tiny component. To understand the motivation
and design goals behind the project read [Why rfw?](./guide/features.md).
