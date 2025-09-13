# API Overview

The RFW framework exposes a collection of focused packages. Each package
encapsulates a domain of functionality, allowing applications to pull in
only the pieces they need. This overview outlines the areas covered by the
public API surface and how the packages fit together.

Core runtime behaviour is provided by packages such as
[core](core), [dom](dom), [events](events) and
[state](state). They implement the rendering loop, DOM bindings and
reactive primitives that form the foundation of every RFW application.

Application level utilities build on top of this foundation. The
[router](router) package offers client‑side navigation, while
[http](http), [assets](assets), [js](js) and [markdown](markdown) simplify
communication with servers, JavaScript interop and markdown rendering.

RFW is designed to run in multiple environments. The [host](host) and
[hostclient](hostclient) packages integrate the runtime with external
systems, and [plugins](plugins) lets developers extend the framework
with custom features. The [docs plugin](docs-plugin) powers this site
by loading markdown files and emitting events for navigation. The
[bundler plugin](bundler-plugin) bundles and minifies JavaScript and
CSS assets in the `static/` directory and inline snippets in RTML
templates during builds.

More specialised modules cover advanced scenarios. Packages like
[animation](animation), [cinema](cinema) and
[webgl](webgl) demonstrate how to drive media playback or GPU
rendering using the same reactive model.

The full set of API references is listed below:

- [animation](animation)
- [assets](assets)
- [cinema](cinema)
- [core](core)
- [components](components)
- [dom](dom)
- [events](events)
- [host](host)
- [hostclient](hostclient)
- [http](http)
- [input](input)
- [js](js)
- [markdown](markdown)
- [shims](shims)
- [plugins](plugins)
- [docs plugin](docs-plugin)
- [highlight](highlight)
- [bundler plugin](bundler-plugin)
- [i18n](i18n)
- [router](router)
- [state](state)
- [webgl](webgl)
- [math](math)
- [game loop](game-loop)
- [netcode](netcode)
- [pathfinding](pathfinding)
