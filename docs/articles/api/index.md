# API Overview

The rfw v2 framework exposes a collection of focused packages. Each package encapsulates a domain of functionality, allowing applications to pull in only the pieces they need.

Core runtime behaviour is provided by [composition](composition), [core](core), [dom](dom), [events](events), and [state](state). They implement the rendering loop, DOM bindings, and reactive primitives that form the foundation of every rfw application.

Application-level utilities build on top of this foundation. The [router](router) package offers client-side navigation with `Page()`, `Group()`, and `Singleton()` helpers.

The [rtmlast](rtmlast) package provides the RTML template parser and renderer. The [host](host) and [hostclient](hostclient) packages integrate the runtime with external systems for SSC mode.

The full set of API references is listed below:

- [composition](composition), type-based component creation, signals, stores, DI, element builders
- [core](core), HTMLComponent, lifecycle, component registry
- [dom](dom), DOM helpers, event delegation
- [events](events), browser event utilities
- [router](router), client-side routing with Page, Group, Singleton
- [rtmlast](rtmlast), RTML template parser and renderer
- [state](state), signals, stores, computed values, watchers
- [signal](signal), Signal[T] reactive values