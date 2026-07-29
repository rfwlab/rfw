# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Breaking change policy

While rfw is pre-stable (v2.0.0 betas), breaking changes may land in any beta
release. Every breaking change is flagged **breaking** in this changelog and
ships with migration notes. Once v2.0.0 stable is released, breaking changes
follow semver: they only land in a new major version.

## [Unreleased]

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

[Unreleased]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.16...HEAD
[2.0.0-beta.16]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.15...v2.0.0-beta.16
[2.0.0-beta.8]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.7...v2.0.0-beta.8
[2.0.0-beta.7]: https://github.com/rfwlab/rfw/releases/tag/v2.0.0-beta.7
