# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Breaking change policy

Stable releases follow semver. Breaking API changes only land in a new major
version and include migration notes.

## [Unreleased]

### Fixed

- resolve `@store`, `@rawstore`, writable bindings and conditions whose module,
  store or key contains a hyphen. `@for` split those paths by hand and accepted
  them, while the other directives matched `\w` and left themselves in the
  rendered markup, so the binding silently showed its own source text.

## [2.2.0] - 2026-08-17

### Added

- `host.NewMuxFS`, an `fs.FS` variant of `host.NewMux`, so an application can
  serve an embedded client build (`go:embed`) and ship as a single self-contained
  binary instead of shipping the build directory alongside it.
- `js.Guard` and `js.OnRuntimePanic` for named recovery boundaries in long-lived
  WebAssembly runtime loops.
- browser-test helpers for asserting active-element state, live input values,
  and stable node identity across reactive renders.

### Changed

- reactive component and `Suspense` updates are coalesced per JavaScript turn
  and committed parent-first through one render scheduler. Existing templates
  remain compatible; keyed list rows opt into move-stable identity with
  `[key expression]` or `data-key`.
- `@for`, `@if`, and ordinary component refreshes now share the same incremental
  reconciler. The parent owns child-root placement, while nested components and
  router outlets retain ownership of their DOM descendants.

### Fixed

- revalidate `rfw_config.js` in production so a deployment cannot leave the
  browser pointing at an older immutable WASM build.
- preserve keyed `@for` row nodes while patching changed content, inserting,
  removing or reordering rows. Conditional blocks no longer replace their DOM
  while the selected branch stays the same, and loops in hidden branches wait
  until that branch becomes visible instead of repainting the mounted view.
- keep the WebAssembly runtime alive after a panic in state listeners, watchers,
  computed values, effects, host push handlers, DOM updates, HTTP observers,
  exposed JavaScript functions, router popstate handling, plugins, or the WASM
  loader. Recovered panics retain their original Go stack and flow through
  `core.ReportError`; one failing listener no longer blocks healthy listeners.
- avoid reading nullable selection offsets from focused number inputs while
  patching the DOM, which previously panicked and terminated the WASM instance.
- preserve the current caret and selection direction when render work briefly
  blurs a focused text field before a DOM patch.
- preserve node identity and live `value`, `checked`, selection, and caret state
  during compatible patches; IME composition no longer publishes partial input
  values to writable store or signal bindings.
- validate a complete DOM patch plan before committing it, so duplicate keys,
  incompatible keyed nodes, and ownership conflicts leave the rendered tree
  untouched.
- preserve handler-owned attributes until the template changes the same
  attribute, preventing unrelated renders from erasing live class and ARIA
  state.
- detect an unexpected exit in the generated JavaScript loader, reload once,
  and show a persistent diagnostic with a manual reload action if the new
  runtime exits again within the recovery window.
- prevent a global delegated `@on` handler from running once per nested
  component root for the same DOM event.
- enable the `pages` plugin by default when `rfw.json` omits the `"plugins"`
  key, so a scaffolded project and the examples register their file-based routes
  without extra configuration. An explicit `"plugins"` block, including an empty
  one, still opts out.

## [2.1.0] - 2026-07-31

### Added

- reactive `state.Batch`, `state.Memo`, `state.Untracked`, and cancellable
  `state.Resource` APIs with keyed request deduplication, cache expiry, reload,
  mutation, and `Suspense` integration.
- component lifecycle scopes for context cancellation, cleanup functions,
  background work, and effects that stop on unmount.
- component-owned event handlers and RTML modifiers for capture, once,
  passive, prevent, stop, self, keyboard filters, exact system keys, debounce,
  and throttle.
- typed SSC actions and forms with strict request decoding, correlated
  responses, field errors, action authorization, handler deadlines, reactive
  connection state, and typed client calls.
- SSC message authorization, session initialization, frame, connection, and
  retained-session limits, per-session rate limits, ordered delivery,
  acknowledgements, replay, and expiring session resume tokens.
- named routes, URL generation, redirects, route metadata, cancellable data
  loaders, reactive navigation state, and browser scroll restoration.
- `Portal`, `KeepAlive`, `Transition`, component DOM lifecycle hooks, and the
  `testkit` package for render, browser interaction, and eventual assertions.

### Changed

- every component instance now receives a distinct ID, including repeated
  instances with the same name and props.
- delegated handlers resolve component-owned registrations before the legacy
  global registry, so repeated components may use the same handler names.
- `Suspense` now patches itself when a tracked resource completes and escapes
  rendered error messages.
- SSC writes are serialized per WebSocket connection and every protocol
  message carries sequence and acknowledgement metadata.
- `host.BindSessionConnection` no longer replaces an active session connection.
  `host.SendSessionOutbound` binds the first connection for a new session;
  resumed sessions require `ReplaySession` or `BindSessionConnection` before
  sending.

### Fixed

- default slot fallback content is preserved when a parent does not provide
  the slot.
- component handler, input binding, effect, context, timer, and DOM hook
  cleanup now runs with the owning component lifecycle.
- resumed sessions reject sends from stale WebSocket connections without
  consuming delivery sequence numbers.

## [2.0.0-beta.19] - 2026-07-30

### Added

- `router.Replace` changes the current browser history entry without adding a
  new one.

### Changed

- generated clients use a content hash in the wasm URL and instantiate the
  module with `WebAssembly.instantiateStreaming` when available. Versioned wasm
  responses from `host` and `ssc` are cached as immutable; plain URLs are
  revalidated.
- nested routes match their resolved full path, including absolute children.
  Parent guards and route parameters are preserved.
- `Suspense` and `ErrorBoundary` now implement the complete wasm `Component`
  lifecycle contract.
- bundled examples and the TodoMVC benchmark declare the beta.19 dependency
  set.

### Fixed

- incremental `@for` patches release replaced input listeners and restore
  legacy and AST store/signal bindings, including after the list becomes empty.
- initial routing, `popstate`, and guard fallbacks no longer add unintended
  browser history entries.
- the wasm loader now stops its progress indicator when module instantiation
  fails.

## [2.0.0-beta.18] - 2026-07-29

### Changed

- the framework's own callbacks (`events.On`/`Once`/`Listen`, the observers,
  `dom.Element.On`, delegated handlers, `http.Request`/`RequestBytes`,
  `js.SetTimeout`, `js.OnAnimationFrame`) now go through `js.SafeFuncOf`. A
  panicking handler aborted the wasm instance and froze the page; it is now
  recovered and routed to `js.OnFuncPanic`. Applications no longer have to wrap
  their own handlers to get the guard.

## [2.0.0-beta.17] - 2026-07-29

### Changed

- **breaking (build):** the module now depends on `go-foundation/v2`. The
  packages rfw uses moved from `pkg/<name>` to `core/<name>` (and `pkg/di` to
  `app/di`); the APIs rfw touches are unchanged. An application pinning
  go-foundation v1 has to bump with it.

## [2.0.0-beta.16] - 2026-07-29

### Added

- `events.Once` attaches a one-shot listener that removes itself and releases
  its callback when the event lands, so FileReader/Image style callbacks no
  longer push `js.FuncOf` and manual `Release` onto application code.
- `dom.From` wraps a raw value into `dom.Element`, the missing entry point for
  elements arriving from an event target or a NodeList.
- `dom.Element.RemoveAttr` and `dom.Element.Matches`, the counterparts of
  `SetAttr` and `Closest`.
- `js.FileReader`, `js.FormData` and `js.WebSocket` constructor accessors, plus
  `js.OnAnimationFrame` for scheduling a plain Go func on the next frame.
- `http.RequestOptions.BodyValue` carries a browser-side body (FormData, Blob,
  ArrayBuffer) for multipart uploads, which the string `Body` cannot express.

### Changed

- a store-driven `@for` now patches its own rows instead of re-rendering the
  whole component: the loop marks its rows with a `data-for` id and inserts
  them after a `<template data-for-anchor>` anchor. A shell whose nav list
  changes on every navigation no longer repaints its dependencies and the
  routed page under it (measured on a real app: ~500ms down to ~20ms per
  navigation). Bodies that pull in components, open their own conditional or
  render more than one root per row still take the full render.

### Fixed

- a conditional branch that was hidden while the state it binds to changed came
  back showing the value it had when it left the DOM; a branch carrying
  bindings now returns through a render.
- two `@if` blocks carrying the same condition shared one id, so they shared
  one content entry and a store change patched the second block's markup into
  the first (an app shell gating its sidebar and its top bar on the same flag
  lost the sidebar). Conditional ids are now positional within the render.
- an `@if` on a store key never reacted when the component had no other
  binding on that store: the condition now subscribes to the keys it reads.
- a routed page vanished, and any DOM it had built for itself after mount was
  wiped, when the shell around its outlet re-rendered (a store-driven `@for`
  or `@if` in a `MountRoot` shell). The outlet subtree is now the router's:
  the patcher leaves it alone and the outlet repaints its child if it is lost.
- an included component was frozen by the parent's render cache, which keys on
  props and dependency identity but not on the store state a template binds
  to: a dependency gated on a shared store key never followed it.
- `dom.DelegateEvents` stacked a second set of listeners when a component was
  delegated twice, firing every handler twice and leaking a `js.Func` per
  event; it now replaces the previous set, and `RemoveDelegatedEvents`
  releases callbacks even when the root is already gone.
- `http.Request` released only the callback that fired and leaked the other
  two per request; both outcomes now release all three, and a rejected body
  promise no longer leaks silently.
- `http.RequestBytes` reached for `js.Global().Call("fetch", ...)` instead of
  `js.Fetch`.

### Changed

- **breaking:** the `ai/pathfinding`, `game`, `webgl`, `netcode` and `animation` packages moved to their own repositories: `github.com/rfwlab/rfw-ai`, `rfw-game`, `rfw-webgl`, `rfw-netcode`, `rfw-animation`. Update imports from `github.com/rfwlab/rfw/v2/<pkg>` to `github.com/rfwlab/rfw-<pkg>`; the APIs are unchanged. The router analytics prefetcher now talks to `hostclient` directly (same wire format).

## [2.0.0-beta.8] - 2026-07-12

Developer-experience release: event delegation works end to end for runtime
markup, binary fetch and input helpers land in the standard APIs, and the
module builds cleanly for wasm.

### Added

- `dom.RegisterHandlerElem` registers delegated handlers that receive the
  element carrying the `data-on-*` attribute, so rows injected with `SetHTML`
  are live without manual re-binding.
- `dom.ExpandEvents` expands the `@on:event:handler` shorthand in markup built
  from Go code; the rtml parser shares the same code path.
- `http.RequestBytes` fetches a response as raw bytes (via arrayBuffer), with
  `js.CopyBytesToGo` exposed for custom interop.
- Element helpers on `dom.Element`: `Val`, `SetValue`, `Checked`, `Data`,
  `Closest`.

### Fixed

- Delegated handlers now receive the resolved element as the second argument;
  previously `evt.target` could point at a child node and `data-*` lookups
  came back empty.
- The wasm loader no longer references its arrayBuffer callback before
  declaration, fixing the js/wasm build of `wasmloader`.
- The `rfw` CLI is excluded from js builds, so
  `GOOS=js GOARCH=wasm go build ./...` succeeds across the module.

### Removed

- Legacy loop paths in the rtml parser (`@foreach` and the dead
  `legacyReplaceForPlaceholders`).

### Docs

- New dynamic-lists guide under `docs/articles/`, fixing the broken README
  link.

## [2.0.0-beta.7] - 2026-07-10

### Fixed

- **breaking:** `@store`, `@prop` and `@for` field substitutions are now
  HTML-escaped by default. Opt into trusted markup with the new `@rawstore:` /
  `@rawprop:` directives. Migration: any template that intentionally binds
  trusted HTML must switch that binding to `@rawstore:`/`@rawprop:`; plain
  text bindings need no change.
- `@for` over an unset store key renders nothing instead of leaking the raw
  loop template into the DOM.
- Type-aware child patching: whitespace-only text nodes are skipped when
  diffing, and unkeyed nodes only patch in place when node name and
  `data-condition` identity match. Fixes duplicated static siblings after
  `@endfor` and sibling `@if` blocks swapping content.
- `Query`/`ByID`/`QueryAll` return a null element when the document is
  unavailable instead of panicking, making the wasm test suite runnable
  headlessly.

[Unreleased]: https://github.com/rfwlab/rfw/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.19...v2.1.0
[2.0.0-beta.19]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.18...v2.0.0-beta.19
[2.0.0-beta.18]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.17...v2.0.0-beta.18
[2.0.0-beta.17]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.16...v2.0.0-beta.17
[2.0.0-beta.16]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.15...v2.0.0-beta.16
[2.0.0-beta.8]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.7...v2.0.0-beta.8
[2.0.0-beta.7]: https://github.com/rfwlab/rfw/releases/tag/v2.0.0-beta.7
