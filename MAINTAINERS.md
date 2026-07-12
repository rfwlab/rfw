# Maintainers

## Project lead

- Mirko Brombin ([@mirkobrombin](https://github.com/mirkobrombin))

The project lead has final say on scope, releases, and API decisions until a
formal governance model is needed.

## Module ownership

| Area | Packages | Owner |
| --- | --- | --- |
| Core runtime | `core`, `composition` | @mirkobrombin |
| DOM and events | `dom`, `events` | @mirkobrombin |
| Router | `router` | @mirkobrombin |
| State | `state` | @mirkobrombin |
| HTTP | `http` | @mirkobrombin |
| JS interop | `js`, `wasmloader` | @mirkobrombin |
| SSC | `ssc`, `host`, `hostclient` | @mirkobrombin |
| CLI | `cmd/rfw`, `plugins` | @mirkobrombin |
| Docs | `docs` | @mirkobrombin |

## Help wanted: co-maintainers

rfw is a single-maintainer project and that is a bus factor of one. Co-
maintainers are welcome for any area in the table above.

A co-maintainer of an area:

- Reviews and merges PRs touching that area.
- Triages issues filed against it.
- Keeps its docs under `docs/articles/` accurate.
- Is consulted on breaking changes affecting it.

There is no minimum time commitment, but consistent responsiveness matters
more than volume. The usual path is: contribute a few meaningful PRs to an
area, then volunteer.

To volunteer, open an issue titled `maintainer: <area>` describing which
packages you want to co-own and pointing to your prior contributions. The
project lead will follow up there.
