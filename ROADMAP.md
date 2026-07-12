# Roadmap

## Current status

rfw is at **v2.0.0-beta.8**. The API surface is close to final; remaining work is
stabilization, not expansion.

## Scope freeze

As of beta.8, **no new features land in core until v2.0.0 stable**. Only bug
fixes, stability work, performance polish, tests, and documentation are
accepted. Feature proposals are welcome as issues but will be scheduled for
post-stable. This is project policy, not a temporary pause.

## Kill-list: modules leaving core

The following packages are out of scope for rfw core and are being extracted
into separate repositories before stable:

- `ai/pathfinding`
- `game`
- `webgl`
- `netcode`
- `animation`

They will keep living under the rfwlab organization but release on their own
cadence and do not block v2.0.0.

## Milestones

### Stability and CI (Q3 2026)

- Green `go test ./...` and `GOOS=js GOARCH=wasm go build ./...` enforced in CI
  on every push.
- Fuzz/edge-case coverage for the rtml parser and the DOM patcher (keyed lists,
  conditionals, escaping).
- SSC reconnect and host/client lifecycle hardening.
- Complete the kill-list extractions.

### Developer experience and docs (Q3 to Q4 2026)

- Guides covering every core package under `docs/articles/` (routing, state,
  SSC, plugins, testing).
- `rfw` CLI polish: clearer errors, stable flags, documented environment
  variables.
- Migration notes for every breaking change shipped during the beta cycle
  (tracked in CHANGELOG.md).

### Ecosystem (Q4 2026)

- Publish the extracted modules as standalone repos with their own docs.
- Project templates beyond the default `rfw init` scaffold.
- Plugin authoring guide and a stable build-plugin interface.

### v2.0.0 stable (target: Q4 2026)

- No known regressions, docs complete, semver guarantees begin.

## Scope boundaries

rfw core will never include:

- Game-engine features (rendering pipelines, physics, pathfinding, netcode).
- A JavaScript component ecosystem or Node-based tooling.
- An ORM, database layer, or backend framework features beyond what SSC needs.
- CSS frameworks beyond the existing build-time Tailwind integration.

Core stays focused: components, rtml templating, reactive state, routing, DOM
interop, HTTP, and Server Side Computed synchronization.
