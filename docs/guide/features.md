# Why rfw?

rfw focuses on delivering a productive Go-first workflow for the web.

- **No virtual DOM** – updates go directly to affected nodes without an
  intermediate tree diff.
- **Written in Go** – reuse Go tooling and packages on both client and
  server.
- **Reactive stores** – computed values and watchers keep state in sync
  with minimal boilerplate.
- **Tiny runtime** – only the code you import ships to the browser and
  very little JavaScript is generated.
- **Extensible** – plugins can augment the compiler or runtime, enabling
  custom build steps or browser integrations.
