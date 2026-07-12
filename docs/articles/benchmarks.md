# Benchmarks: rfw vs Svelte, Solid, and htmx

"Fine-grained reactivity" and "no virtual DOM" are claims. This page puts
numbers next to them. The numbers are not flattering everywhere (rfw ships
a WebAssembly binary that is two orders of magnitude larger than a compiled
Svelte or Solid bundle) and they are published anyway, because numbers,
even bad ones, beat adjectives.

Everything here is reproducible from the repository: the four apps live
under [`bench/todomvc/`](../../bench/todomvc/), the harness is
[`bench/measure.mjs`](../../bench/measure.mjs), and the raw output is
[`bench/results.json`](../../bench/results.json).

## What was measured

Four equivalent TodoMVC implementations (add todo, toggle done, delete,
filter all/active/done, active-items counter), plus a live-update scenario:
a counter element updated 10 times per second for 3 seconds (30 ticks).

| Implementation | Stack |
| --- | --- |
| **rfw** | Go 1.26 compiled to wasm, rfw `v2.0.0-beta.8` (local checkout), built with `GOOS=js GOARCH=wasm go build -ldflags="-s -w"` |
| **Svelte** | Svelte 5.56.4 (runes), Vite 6.4.3 production build |
| **Solid** | solid-js 1.9.14, Vite 6.4.3 production build |
| **htmx** | htmx 2.0.10 served locally plus a small Go `net/http` server rendering HTML fragments for add/toggle/delete/filter, idiomatic `hx-trigger="every 100ms"` polling for the live counter |

Metrics, per framework, 3 runs each, medians reported:

- **Load**: navigation start to app interactive, defined as the moment a
  `requestAnimationFrame` poll (installed before navigation) sees both the
  new-todo input and the todo list in the DOM.
- **Add 100 todos**: 100 todos added through the UI with real DOM events
  (native value setter, `input` event, button click), waiting for each
  row to appear before the next add. Executed in-page so CDP round-trips
  do not pollute the measurement.
- **Live update**: click a start button, then 30 counter updates at 10 Hz.
  Expected wall time: 3000 ms. A `MutationObserver` counts distinct
  rendered values, so dropped updates would show up as a count below 30.
- **Memory**: `JSHeapUsedSize` via CDP after load, plus the WebAssembly
  linear memory size for the rfw app (wasm memory is *not* included in the
  JS heap number).
- **Bundle size**: every byte the browser needs to boot the app, raw and
  `gzip -9`.

### Machine caveat

Single Linux laptop (Chrome 150, headless, Node 24, Go 1.26), no CPU
pinning, no isolation. Load times on localhost fluctuated noticeably
between runs (tens of milliseconds). **Relative numbers matter more than
absolute ones.** Run the harness on your own hardware before quoting any
figure.

## Results

Bundle size, meaning everything the browser downloads to boot the app:

| Framework | Raw | gzip -9 | What's in it |
| --- | ---: | ---: | --- |
| rfw | 8,906 KB | 2,430 KB | `app.wasm` 8,886 KB (2,424 KB gz) + `wasm_exec.js` + loader + html |
| Svelte | 38 KB | 15 KB | one JS chunk + html |
| Solid | 15 KB | 6 KB | one JS chunk + html |
| htmx | 51 KB | 17 KB | `htmx.min.js` + server-rendered initial page |

The unoptimized rfw build (`go build` without `-ldflags="-s -w"`) is
9,067 KB raw / 2,467 KB gzip. Stripping symbols saves surprisingly little;
the weight is the Go runtime and stdlib, not debug info.

Runtime (medians of 3 runs):

| Framework | Load (ms) | Add 100 todos (ms) | Live update (ms, 3000 expected) | Updates rendered (30 expected) | JS heap (MB) | wasm memory (MB) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| rfw | 379 | 490 | 3010 | 30 | 2.2 | 38.5 |
| Svelte | 98 | 478 | 3009 | 30 | 1.9 | n/a |
| Solid | 75 | 45 | 3008 | 30 | 1.8 | n/a |
| htmx | 77 | 2985 | 3886 | 30 | 2.0 | n/a |

Lines of application code, for scale: rfw component 148 lines of Go,
Svelte 76 lines, Solid 82 lines, htmx server 187 lines of Go (the htmx
number includes all the fragment rendering; with htmx the "framework" is
your server).

## Honest analysis

**Bundle size: rfw loses, badly, as expected.** 2.4 MB gzipped versus 6 KB
for Solid is not "an order of magnitude", it is closer to two and a half.
A Go wasm binary carries the Go runtime, garbage collector, and every
stdlib package the app touches. TinyGo would shrink this substantially but
is not the supported toolchain today. If first-visit bundle size on a slow
connection is your top constraint, rfw is the wrong tool and this table
says so plainly.

**Load time: rfw is roughly 4-5x slower to interactive on localhost.**
About 380 ms versus 75-100 ms. That is fetch plus compile plus instantiate
of a 9 MB wasm module plus Go runtime startup. On localhost the fetch is
nearly free; over a real network the gap grows with the download. Caching
and `Cache-Control: immutable` make repeat visits cheap, but the first
visit pays full price.

**Interaction: rfw keeps up with Svelte, Solid is in another league.**
Adding 100 todos through real UI events took rfw about 490 ms and Svelte 5
about 478 ms, statistically the same on this machine. Solid's 45 ms is the
fine-grained-reactivity headline number: it only appends one row per add,
while the rfw implementation re-renders the visible list on each change
(the idiomatic `SetHTML` pattern) and Svelte pays proxy/effect overhead.
The Go/JS boundary crossing does not make rfw an outlier here; wasm is
slow to load, not slow to run. htmx's roughly 3 s is the cost of its
architecture: one HTTP round-trip and one list re-render per add, measured
on localhost. Add real network latency and it scales linearly with it.

**Live update: everyone rendered all 30 frames.** rfw, Svelte, and Solid
all finished within about 10 ms of the expected 3000 ms with zero dropped
updates. A 10 Hz store-driven counter is trivial for all three, and rfw's
store binding updates the DOM as promptly as the compiled JS frameworks.
htmx took about 3.9 s for the same 30 updates: polling re-arms 100 ms after
each response, so per-tick overhead accumulates. This is the quantified
version of the comparison doc's claim that client-side state beats
request-per-update for high-frequency changes, and it would get worse for
htmx over a real network, while the rfw/Svelte/Solid numbers would not
change at all.

**Memory: the honest number for rfw is about 40 MB, not 2 MB.** The JS
heap looks tiny (2.2 MB) but the Go runtime sits in wasm linear memory,
which grew to about 38 MB after load and 100 todos. Svelte and Solid do
everything in under 2 MB of JS heap. For a desktop browser this is
irrelevant; for a low-end phone with many tabs it is not.

## What rfw buys for that price

The table above is what rfw pays. This is what it gets:

- **One language.** State, handlers, rendering, and the server share Go.
  No context switch, no serialization layer to design, no TypeScript
  mirror types drifting from the backend structs.
- **Type safety end to end.** The compiler checks the todo struct, the
  handler signatures, and the store access.
- **No npm supply chain.** The rfw app's full dependency tree is the Go
  module cache, verified by `go.sum`. The Svelte and Solid builds pulled
  39 and 72 npm packages respectively, each one a trust decision.
- **Predictable updates.** The live-update numbers show rfw's store
  bindings are not a bottleneck: once loaded, the app is as responsive as
  compiled-JS frameworks for this workload.

If your product is a content site where every first-visit millisecond
converts, use Svelte or Solid; the numbers are unambiguous. If it is an
app behind a login where the bundle is downloaded once and cached, the
load-time gap collapses into a one-time cost and the interaction numbers,
which are competitive, are what your users actually feel.

## Reproducing

```sh
# rfw app
cd bench/todomvc/rfw
GOOS=js GOARCH=wasm go build -o app.wasm .
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o app.opt.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Svelte / Solid
cd ../svelte && npm install && npm run build
cd ../solid  && npm install && npm run build

# htmx
cd ../htmx
npm install --no-save htmx.org && cp node_modules/htmx.org/dist/htmx.min.js .
go build -o htmx-server server.go

# measure (needs Chrome; set CHROME_PATH if not /usr/bin/google-chrome)
cd ../.. && npm install && node measure.mjs
```

Known gaps versus the original issue: Datastar and a templ-based htmx
variant were not implemented (the htmx server uses hand-rolled fragment
rendering instead of templ; same architecture, one less dependency), and
sizes are reported as gzip rather than brotli. The live-update scenario
uses htmx polling, not WebSockets; a WebSocket variant would narrow htmx's
live-update gap at the cost of more plumbing.
