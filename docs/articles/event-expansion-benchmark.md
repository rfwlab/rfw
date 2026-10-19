# Event expansion allocation measurement

`dom.ExpandEvents` keeps literal data separate from event attributes using the
HTML tokenizer. This measurement compares the safe implementation at
`3cf41b809766ded33fa375656cdb0161d177f0c1` with direct ASCII directive parsing,
builder writes and a reused per-call attribute-offset slice. Tokenizer states
and original-byte slicing are unchanged. No shared buffer or dependency is added.

## Local results

Measured on 2026-10-17, Linux amd64, Go 1.26.9, GOMAXPROCS=4. Each fixture has
1,000 rows; input construction is outside the benchmark loop. Fixtures are in
`dom/markup_context_test.go`: `BenchmarkExpandEventsRows` includes SSH URLs and
literal directive-shaped text; `BenchmarkExpandEventsDirectiveRows` contains
ordinary issue-row markup with a click directive.

Five 300ms samples per fixture were run for parent then candidate, followed by
five candidate then parent samples. The table combines ten samples per fixture.
These are local synthetic parser costs, not application latency or throughput.

| Fixture | Source | Median ns/op | Range ns/op | Median B/op | Median allocs/op |
| --- | --- | ---: | ---: | ---: | ---: |
| Mixed literals | Parent | 1,200,066.5 | 1,119,634-1,282,948 | 396,366.5 | 8,011 |
| Mixed literals | Candidate | 707,277 | 680,121-731,147 | 282,097 | 4,010 |
| Directive rows | Parent | 1,208,447.5 | 1,146,223-1,246,741 | 354,964 | 8,010 |
| Directive rows | Candidate | 664,982 | 644,224-680,930 | 241,136 | 4,010 |

Median time fell 41.1% for mixed literals and 45.0% for directive rows on this
machine. Allocation counts fell about 50%; allocated bytes fell 28.8% and 32.1%.
Parent allocation bytes varied slightly across samples with runtime bookkeeping.
No confidence interval, low-power device, browser timing or application-wide
speed improvement is claimed. Compiler and hardware changes require remeasurement.

## Reproduction

Use two clean checkouts: the parent above and the revision containing this
article. With the same Go installation and environment, run in each checkout:

```sh
GOMAXPROCS=4 go test ./dom -run '^$' -bench BenchmarkExpandEvents \
  -benchmem -benchtime=300ms -count=5
```

Run parent then candidate, then reverse the order. Compare medians and ranges,
not the best sample. For this recorded run, Go overlays substituted the parent's
`dom/markup.go` while keeping fixture source and toolchain identical. The new
grammar test was excluded only for the parent, which lacks the private parser.
The regular parent context tests remain available at the cited revision.

For profiling, use `-cpuprofile cpu.pprof -memprofile mem.pprof` in a separate
run. The original safe parser spends time in the tokenizer, regexp matching and
formatting; the optimization removes regexp/formatting work without dropping
the tokenizer's raw-text and comment handling.

## Compatibility checks

`FuzzEventDirectiveMatchesExistingGrammar` compares the private parser with the
previous anchored regular expression, including short `@on:handler`, dotted
modifiers, malformed words and non-ASCII input. A 10-second local run completed
807,069 executions without a mismatch. This is evidence for sampled inputs,
not a proof over all possible strings.

Existing context tests cover quoted/unquoted values, comments, raw-text elements,
unfinished tags, idempotence and whitespace. New regression tests cover adjacent
directives and reuse across tags. The grammar stays ASCII; validated handler and
event words need no HTML quote escaping. Raw-text content is never rescanned as
an event attribute.

This measurement is separate from the historical TodoMVC comparison. It contains
no new Svelte, Solid or htmx measurement and does not establish framework rankings.
