# Hot reload: what is instant, what is not

`rfw dev` watches the project and reacts differently depending on what you
save. This page states plainly what each edit costs, measured on the
`examples/dashboard` app (Go 1.25, amd64 laptop, warm build cache).

## Template edits (.rtml): instant

Saving an `.rtml` file streams the new markup to the browser over the HMR
channel and swaps the component's template in place, before any compilation
happens. Feedback is effectively immediate (tens of milliseconds, dominated
by the file watcher). Component and store state are preserved.

The wasm binary embeds templates, so the dev server still rebuilds in the
background to keep `dist/` consistent; you only pay that cost if you hard
refresh the page.

## Go edits (.go): one warm compile plus a reload

There is no Go equivalent of JavaScript module replacement: the wasm binary
is rebuilt and the page reloads. Measured:

| Step | Time |
| --- | --- |
| Warm rebuild after editing one file | ~370 ms |
| First build after `go clean -cache` | ~8.6 s |

A warm edit-save-reload cycle lands well under a second on the example apps;
real applications pay proportionally to their own code size, not the
framework's, because rfw's packages stay cached.

The page reload resets in-memory component state. Stores using the
persistence helpers (`state/persistence`) survive the reload; plain signals
and component fields do not.

## Failure behavior

A compile error does not kill the dev server: the error is printed, a failed
rebuild event is emitted on the HMR bus, and the watcher keeps running so the
next save picks up cleanly.

## Known limits

- No incremental wasm linking: even a one-line Go change relinks the binary
  (~370 ms warm). This is a Go toolchain property, not something rfw can
  shortcut today.
- The cold build after cache eviction or a toolchain upgrade takes seconds;
  subsequent builds are warm.
- Template hot swap matches templates to components by name; a template used
  by a component the dev server cannot associate falls back to a full page
  reload.
