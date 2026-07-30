# Roadmap

## Current status

rfw 2.1 establishes the application framework APIs around its Server Side
Computed runtime. SSC remains the defining feature: server state, typed
actions, ordered delivery, and browser hydration share one protocol.

## 2.1

The 2.1 line covers the framework foundation needed by production
applications:

- reactive batching, memoized values, asynchronous resources, and component
  lifecycle scopes;
- typed SSC actions and forms, authorization hooks, resource limits, ordered
  delivery, and resumable sessions;
- route loaders, route names, redirects, metadata, and scroll restoration;
- component-owned events, DOM hooks, Portal, KeepAlive, Transition, and a
  browser testkit.

Work after the release is limited to fixes, measurements, and documentation
until 2.2 development begins.

## 2.2

The 2.2 priority is the native rfw component library. It will ship as a
separate package with accessible primitives, design tokens, form controls,
navigation, overlays, data display, and documented extension points. The
library must remain usable without Node or a frontend package manager.

The same cycle will expand:

- testkit event and accessibility assertions;
- SSC operational telemetry and deployment examples;
- router tooling for generated route names;
- official project templates and plugin authoring documentation.

## Scope boundaries

rfw core does not include:

- server-side rendering. SSC is the server-driven rendering and
  synchronization model;
- an ORM or database abstraction. Applications use the Go database and ORM
  packages that fit their backend;
- npm, Node-based builds, or a JavaScript component ecosystem. Prebuilt
  browser libraries can be vendored as static assets and attached through DOM
  lifecycle hooks;
- game-engine features such as rendering pipelines, physics, pathfinding, or
  netcode;
- a CSS framework beyond the existing build-time Tailwind integration.
