# RTML render cost

Working notes on the template engine's render path, written while cutting the
cost of mounting a 118 row issue board. Everything here was measured, not
inferred.

## How to reproduce

The render path is behind `//go:build js && wasm`, so the benchmarks only run
in a browser:

```bash
go install github.com/agnivade/wasmbrowsertest@v0.11.0
GOOS=js GOARCH=wasm go test -p=1 -exec "$(go env GOPATH)/bin/wasmbrowsertest" \
  -run '^$' -bench 'Board|For' -benchmem -benchtime 100x ./core/
```

`wasmbrowsertest` drives chromedp, which looks for `google-chrome` on PATH. Any
Chrome or Chromium works; on a machine without one, a wrapper pointing at the
Playwright cache is enough, and it needs `--no-sandbox` the way CI does:

```sh
#!/bin/sh
exec "$HOME/.cache/ms-playwright/chromium-*/chrome-linux64/chrome" --no-sandbox "$@"
```

The benchmarks live in `core/rtml_bench_test.go`:

- `BenchmarkRenderBoard118`, a full `HTMLComponent.RenderFresh` of a component
  whose template holds a 118 row `@for`, an `@if` with two sibling elements, a
  `@store` binding, an `@expr` and one `@on:click` per row.
- `BenchmarkForExpansion118`, `replaceForPlaceholders` on its own.
- `BenchmarkRenderBoardStatic`, every pass that runs after the loop expands,
  which is where the document is large.

## Numbers

100 iterations, headless Chrome 151, js/wasm.

| benchmark | before | after | factor |
|---|---|---|---|
| `BenchmarkRenderBoard118` | 59.39 ms/op, 4766 kB/op, 18136 allocs/op | 5.01 ms/op, 1283 kB/op, 2403 allocs/op | 11.9x time, 3.7x bytes, 7.5x allocs |
| `BenchmarkForExpansion118` | 8.42 ms/op, 2133 kB/op, 14013 allocs/op | 2.11 ms/op, 701 kB/op, 1919 allocs/op | 4.0x time, 3.0x bytes, 7.3x allocs |
| `BenchmarkRenderBoardStatic` | 41.83 ms/op, 2480 kB/op, 4001 allocs/op | 2.28 ms/op, 579 kB/op, 463 allocs/op | 18.3x time, 4.3x bytes, 8.6x allocs |

## Where the time was

A temporary harness benchmarked each pass on its own against the expanded board
(118 rows, about 130 kB of markup). Roughly, at 20 to 30 iterations, so read
these as proportions rather than exact figures:

| pass | before | after |
|---|---|---|
| `replaceConstructors` | 21.6 ms | 0.01 ms |
| `replaceRtIsAttributes` | 19.1 ms | 0.13 ms |
| `replaceExprInClassAttr` | 8.15 ms | 0.09 ms |
| `replaceEventHandlers` (`dom.ExpandEvents`) | 1.84 ms | 1.00 ms |
| `replaceConditionals` | 1.14 ms | 0.43 ms |
| `minifyInline` | 0.78 ms | 0.09 ms |
| `replaceComponentIncludes` | 0.48 ms | 0.01 ms |
| `replaceStorePlaceholders` | 0.46 ms | 0.08 ms |
| everything else, each | under 0.25 ms | under 0.10 ms |

Three passes were 92 percent of the post-expansion cost, and they share one
trait: their pattern is anchored on a token every element carries (`<` for
`reConditionalAttr` and `reRtIs`, `class="` for the class `@expr` pattern) and
then backtracks over `[^>]*` or `[^"]*`. On a template they run a handful of
times. On the same template after a `@for` expands, they restart at every one of
118 rows worth of elements and find nothing.

That also explains the CPU profile that started this: `regexp.machine.add`,
`regexp.machine.step` and `regexp.Regexp.tryBacktrack` at the top, with the GC
entries underneath them being the copies each pass produced.

## What changed

### `core/rtml.go`, gate every pass on the token it needs

Each `Replace*Placeholders` function now returns early when the template cannot
contain its directive. This is not only about the regexp engine:
`Regexp.ReplaceAllStringFunc` copies the whole document even when it matches
nothing, so a no-match pass cost a 130 kB copy on top of the walk.

- `replaceRtIsAttributes`: `rt-is="`
- `replaceIncludePlaceholders`, `replaceComponentIncludes`: `@include:`
- `extractSlotContents`, `replaceSlotPlaceholders`: `@slot`
- `replaceStorePlaceholders`: `@rawstore:` and `@store:` checked separately, so
  the raw half split out into `replaceRawStorePlaceholders`
- `replaceSignalPlaceholders`: `@signal:`
- `replaceExprPlaceholders`: `@expr:`
- `replacePropPlaceholders`: `@rawprop:` and `@prop:`
- `replacePluginPlaceholders`: `{plugin:` and `@plugin:`
- `replaceConditionals`: `@if:`
- `minifyInline` in `core/html_component.go`: `<script` or `<style`

Two guards needed more than a substring check, because the cheap token is too
common to be useful:

- `hasConstructorToken` scans for a `[` preceded by whitespace, which is what
  `reConditionalAttr` requires before it can match anything.
- `hasClassExpr` walks the `class="` attributes and looks for `@expr:` inside
  one, which is exactly the condition `reClassExpr` tests, using byte scans
  instead of the regexp engine.

`hasClassExpr` is the one place where this style of guard bit back. It first
resumed the scan past the attribute value it had just rejected, which loses a
`class="` that starts inside that value. A `<code>class="</code>` in the page
text was enough to hide the next real class expression from the pass. It now
resumes right after the token. `TestHasClassExprMatchesPattern` pins it to
`reClassExpr` on both shapes, and `TestConstructorTokenImpliedByPattern` pins
the weaker implication `hasConstructorToken` has to satisfy. A guard that is
cheaper than its pattern is only safe while it agrees with it, so any new one
belongs in those tests.

One behavioural detail worth keeping in mind: `replaceConditionals` used to
split the template into lines and join them back with a terminator on each, so
its output always carried a trailing newline. The shortcut adds that newline
back. Removing it breaks `TestGoldenConstructorDirectives`,
`TestGoldenPluginDirectives` and `TestGoldenHostDirectives`.

### `core/rtml.go`, one text node per run of lines

`parseBlock` allocated a `TextNode` per line. On an expanded board that is about
1900 nodes and 3000 allocations to rebuild a string that came in as one piece.
It now accumulates runs of plain lines and emits one node per run. Output is
byte for byte the same.

### `core/rtml_for.go`, the per row path

- `reFor` is a package variable. It used to be compiled inside
  `replaceForPlaceholders`, once per render pass.
- `aliasRegexps` caches the per alias `@prop` and `@rawprop` matchers in a
  `sync.Map`, with `regexp.QuoteMeta` on the alias. `substituteItemFields` runs
  once per item, so a 118 row list compiled 236 patterns per render.
- `substituteFieldRefs` replaces the `ReplaceAllStringFunc` plus
  `FindStringSubmatch` pair with `FindAllStringSubmatchIndex` and a
  `strings.Builder`. The old shape ran each pattern twice per match, once to
  find it and once inside the callback to reach the capture group.
- `resolveNestedKey` takes a direct map lookup when the field path has no dot,
  which is the common case and skips a `strings.Split` per field.

The first two items are the patch from `perf/rfw-regex-cache`, ported as they
were. On the app they had measured no improvement, and that reading was correct
at the time: they were worth about 6 ms against 45 ms of pass cost that hid
them. With the passes gated, `BenchmarkForExpansion118` moves from about 10 ms
to about 3.7 ms on their own.

### `core/rtml.go`, `escapeValue`

Takes the value directly when it is already a string, instead of going through
`fmt.Sprintf("%v", v)`. Every directive substitution goes through this function.

### `dom/markup.go`, `ExpandEvents`

Same rewrite as `substituteFieldRefs`: match indexes and a builder instead of
`ReplaceAllStringFunc` with a second execution inside the callback. A board
carries one `@on:` directive per row, so this pass does real work; it went from
1.84 ms and 843 allocations to 1.00 ms and 361.

## Tests

`core/rtml_test.go` gained three cases that lock the behaviour the changes had
to preserve. All three pass on the pre-change tree as well, which is the point:

- `TestConditionalWrapsSiblingsInOneContainer`, an `@if` body holding two
  sibling elements renders both inside one `data-condition` container, in
  order. The wrapper is load bearing and callers style around it.
- `TestForBodyConditionalPerRow`, an `@if` inside a `@for` body gets its own
  conditional block per row, so one row's branch cannot patch another's.
- `TestForBodyWithoutDirectives`, a plain loop body still renders one row per
  item.

`go test ./...` and the js/wasm suite are green before and after, as is
`golangci-lint run` on both targets.

Beyond the suite, a throwaway harness rendered a 16 template corpus covering
every directive (`@for` over a slice, a map and a range; `@if` / `@else-if` /
`@else` with sibling and nested bodies; `@store`, `@rawstore`, writable
bindings, `@signal`, `@expr` in text and in a class list, `@prop`, `@rawprop`,
`{{prop}}`, `@plugin`, `@on:` with modifiers, `[ref]` and `[key]`
constructors, a Tailwind class list with brackets, inline `<script>` and
`<style>`, and a template with no directive at all). Its output is byte for
byte the same on `main` and on this branch. That harness was not kept: it is
twenty lines and easier to rewrite than to maintain.

## Deliberately not changed

**The single pass scanner.** This was the first option on the list and the
numbers stopped justifying it. After gating, the whole post-expansion pipeline
is 2.28 ms of a 5.01 ms render, and its largest single item is
`dom.ExpandEvents` at 1.00 ms doing real substitution on 118 real matches.
Replacing fourteen functions, and every escaping decision inside them, to
compete for a fraction of 2.28 ms is a poor trade against the risk.

**The parsed template cache.** It would not help. What remains is dominated by
`replaceForPlaceholders` at 2.11 ms, which depends on the collection, not on the
template. Component templates are already handed in as `//go:embed` byte slices
and only parsed into nodes by `replaceConditionals`, which is now 0.43 ms. A
content hashed parse tree would still have to re-run every substitution on every
render.

**`HTMLComponent.Render`'s cache.** `cacheKey` hashes props and dependency
identity, never the store state a template binds to, which is why `RenderFresh`
exists to bypass it. Left alone.

**`reConditionalAttr` still misreads some Tailwind class lists.** A class list
like `class="flex [&>*]:mt-2"` puts whitespace before a bracket, which is all
the constructor pattern needs, and it will emit `data-ref="&>*"`. This predates
the changes here and `hasConstructorToken` does not make it worse, it only skips
the pass when no such pair exists at all. Worth a separate fix.

**`insertRowMarkers` and `insertDataKey`** still build their attributes with
`fmt.Sprintf` per row. Measured at well under 0.1 ms for 118 rows.

## What is left

At 118 rows a fresh render is 5.01 ms, split roughly as 2.11 ms of loop
expansion, 2.28 ms of passes over the expanded markup, and the rest in the
component's own prologue. The next targets in order of size are
`dom.ExpandEvents` (1.00 ms, genuine work), `replaceConditionals` (0.43 ms) and
the remaining per row allocations in `expandForRows`. None of them is worth a
structural rewrite on its own.
