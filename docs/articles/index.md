# rfw documentation

## Guides

- [The mental model](guide/mental-model.md): rfw in a single concept; the
  DOM as a projection of Go state.
- [Why rfw (and why not)](guide/why-rfw.md): an honest comparison with
  templ + HTMX, Datastar, and SPA frameworks.
- [Getting started from Node](guide/getting-started-from-node.md): install Go,
  fix the PATH gotcha, and map npm mental models to the Go toolchain.
- [Build a real-time dashboard in 30 minutes](guide/realtime-dashboard-tutorial.md):
  scaffold, stores, `@for` lists, event handlers, and a live data feed.
- [Dynamic lists and events](guide/dynamic-lists.md): fetch data, render
  rows, handle clicks; the pattern behind every real page.
- [SSC security model](guide/ssc-security.md): what crosses the wire, what
  stays on the server, and the auth work rfw leaves to you.
- [Hot reload: what is instant, what is not](guide/hot-reload.md): measured
  rebuild times and the honest limits of the dev loop.

## Measurements

- [Benchmarks](benchmarks.md): TodoMVC and live updates vs Svelte, Solid and
  htmx; bundle sizes, load, interaction and memory.
