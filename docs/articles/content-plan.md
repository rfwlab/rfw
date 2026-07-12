# 90-day content plan

Audience: Go backend engineers who currently bolt a JS frontend (or
templ/htmx) onto their services and would rather not. Every piece leads with a
working artifact, not a manifesto. All code shown must run against the current
`v2` API; reuse the verified snippets from the guides in this repository.

Cadence: one article per month, screencast in the last month, each piece
cross-posted within the same week it is published.

## Month 1: Article. "A real-time dashboard in Go, no JavaScript, no REST API"

The written version of the dashboard tutorial, tightened for an external
audience.

- Hook: the standard Go stack for an internal dashboard today is a REST API
  plus a React app; count the moving parts, then delete them.
- Scaffold with `rfw init`, show the generated tree, explain that the browser
  runs a wasm binary compiled from your Go.
- Build the metrics store and template bindings (`@store`, `@for`,
  `@on:click`); goroutine + `time.Ticker` as the data feed.
- Honest limits section: wasm binary size, pre-1.0 API, when templ/htmx is
  the better fit.
- Close with the repo link and the 30-minute tutorial for the full version.

Publish: dev.to (canonical), cross-post to r/golang with a comment thread
answering "why not htmx", submit to Hacker News as Show HN with the demo GIF.

## Month 2: Article. "From templ + htmx to rfw: what a state engine buys you"

Comparison piece aimed at the people most likely to adopt: those already doing
server-driven UI in Go.

- Rebuild one identical page three ways: templ+htmx, datastar, rfw; same
  feature, side-by-side source.
- Where htmx is enough: request/response UI, forms, low interactivity. Say so
  plainly; this earns trust.
- Where it breaks down: server-pushed state, multiple widgets sharing state,
  optimistic updates; show the htmx workarounds versus rfw's store sync.
- SSC in one diagram: what runs on the host, what runs in wasm, what the
  WebSocket carries.
- Migration notes: what a templ project keeps (handlers, models) and what it
  drops (fragment endpoints).

Publish: dev.to (canonical), r/golang, Golang Weekly newsletter submission.
HN only if the benchmark/diagram section is strong enough to stand alone.

## Month 3: Article. "Anatomy of an rfw component: what actually happens between Set and the DOM"

Internals piece for the engineers who read the first two and want to know the
cost model before committing.

- Trace one `store.Set("cpu", ...)` end to end: listener fire, binding update,
  `data-store` span patch; with file/line references into the rfw source.
- The template pipeline: RTML directives to AST to rendered HTML, and why
  there is no virtual DOM diff for plain bindings.
- Event delegation: one listener at the component root, `data-on-*`
  resolution, why runtime-injected markup stays live.
- Render caching and `RenderFresh`: when a component re-renders wholesale and
  what that costs.
- What is deliberately not there yet (pre-1.0): the honest gaps list, linked
  to open issues so readers can pick one up.

Publish: dev.to (canonical), r/golang, HN (internals posts about Go+wasm
historically perform well there). Link from the repository README.

## Month 3: Screencast. "Build a real-time dashboard in Go in 30 minutes"

The dashboard tutorial as a single-take video; the article is the script, so
production cost is editing, not writing.

- Format: 25-30 minutes, real terminal and browser, no slides after the first
  minute; chapters matching the tutorial's numbered steps.
- Minute 0-3: the pitch and the finished dashboard running; then `go install`,
  `rfw init`, first page load.
- Minute 3-20: store, component, template, feed goroutine, pause button; type
  everything on camera, let the dev server reload do the demos.
- Minute 20-25: swap the simulated ticker for a real data source to prove the
  loop shape holds; recap the mental model (store in, DOM out).
- Description links: the written tutorial, the repo, the getting-started-from-
  Node guide for viewers without Go installed.

Publish: YouTube (canonical), embed at the top of the dev.to tutorial
article, short clip (the pause-button moment) for r/golang and the HN thread.

## Operational notes

- Each article ends with the same two links: the tutorial and the repo. One
  call to action, not five.
- Reuse the `docs/assets/hero-counter.gif` style: every post needs one moving
  image above the fold.
- Track one metric per channel: dev.to follows, r/golang upvote ratio, HN
  front-page minutes, YouTube average view duration. Review after day 90 and
  cut the weakest channel.
