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

[Unreleased]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.8...HEAD
[2.0.0-beta.8]: https://github.com/rfwlab/rfw/compare/v2.0.0-beta.7...v2.0.0-beta.8
[2.0.0-beta.7]: https://github.com/rfwlab/rfw/releases/tag/v2.0.0-beta.7
