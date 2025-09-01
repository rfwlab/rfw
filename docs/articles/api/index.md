# API Overview

The RFW framework exposes a collection of focused packages. Each package
encapsulates a domain of functionality, allowing applications to pull in
only the pieces they need. This overview outlines the areas covered by the
public API surface and how the packages fit together.

Core runtime behaviour is provided by packages such as
[core](core.md), [dom](dom.md), [events](events.md) and
[state](state.md). They implement the rendering loop, DOM bindings and
reactive primitives that form the foundation of every RFW application.

Application level utilities build on top of this foundation. The
[router](router.md) package offers client‑side navigation, while
[http](http.md) and [js](js.md) simplify communication with servers and
JavaScript interop.

RFW is designed to run in multiple environments. The [host](host.md) and
[hostclient](hostclient.md) packages integrate the runtime with external
systems, and [plugins](plugins.md) lets developers extend the framework
with custom features.

More specialised modules cover advanced scenarios. Packages like
[animation](animation.md), [cinema](cinema.md) and
[webgl](webgl.md) demonstrate how to drive media playback or GPU
rendering using the same reactive model.

The full set of API references is listed below:

- [animation](animation.md)
- [cinema](cinema.md)
- [core](core.md)
- [dom](dom.md)
- [events](events.md)
- [host](host.md)
- [hostclient](hostclient.md)
- [http](http.md)
- [js](js.md)
- [plugins](plugins.md)
- [i18n](i18n.md)
- [router](router.md)
- [state](state.md)
- [webgl](webgl.md)
