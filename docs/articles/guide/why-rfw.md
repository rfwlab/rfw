# Why rfw (and why not)

If you are choosing between rfw, `templ` + HTMX, and Datastar, you are
already sold on server-driven UI. The three differ in how much machinery
they put between your Go code and the browser, and in what they ask you to
write when the page needs real client-side behavior. This page is an honest
comparison; the losses section is as load-bearing as the wins section.

## What each approach is

**templ + HTMX.** `templ` gives you type-checked HTML components in Go;
HTMX makes elements issue HTTP requests and swap the returned fragments
into the page. The unit of exchange is an HTML fragment over a
request/response cycle. There is no persistent state channel: every
interaction is a round trip that re-renders some part of the page on the
server.

**Datastar.** A hypermedia framework built on Server-Sent Events. The
server pushes signal patches and HTML fragments over a long-lived SSE
connection; declarative `data-*` attributes bind them into the page.
It keeps a reactive state model synchronized from the server without a
frontend framework, but client-side expressions live in attribute strings,
outside your compiler.

**rfw.** Your components, state, and event handlers are Go, compiled to a
wasm binary that runs in the browser. Templates bind directly to Go stores
and signals. For server-driven pages, Server Side Computed (SSC) components
keep server state and the DOM synchronized over a persistent WebSocket:
host handlers run in a normal Go server process, per-session stores hold
per-user state, and the client runtime hydrates values, verifies them
against expectation hashes, and requests a resync snapshot on mismatch.

**React/Vue-style SPAs**, for completeness: they solve a different problem,
a rich client application that owns its state, talking to your backend
through an API you also have to design, version, and secure. You get the
largest ecosystem and hiring pool in the industry, at the price of two
codebases, two type systems, and a contract between them that nothing
checks end to end. If your product genuinely is a client-heavy app with
offline needs and a large frontend team, that trade can be right, and rfw
is not trying to win it.

## Where each one wins

**templ + HTMX wins** when the app is mostly pages and forms. It is the
simplest of the three: no persistent connections, plain HTTP semantics,
caching and load balancing work as they always have, and the client-side
payload is a ~14 kB script. It degrades gracefully and is easy to reason
about operationally.

**Datastar wins** when you want server-pushed reactivity with a tiny client
footprint and you are comfortable with SSE and attribute-based expressions.
It gets you live updates without WebSockets and without shipping an
application runtime to the browser.

**rfw wins** when the page itself has real logic:

- **One language, type-safe end to end.** Client and server are the same Go
  module. A host handler and the component consuming it share types; the
  contract between browser and server is checked by the compiler, not by
  discipline. There is no template micro-language for logic either:
  conditions and expressions in `.rtml` resolve against Go values.
- **A state synchronization engine, not just fragment swaps.** SSC gives
  you per-session server-side stores, targeted or broadcast pushes over one
  WebSocket, hydration with hash-verified expectations, and automatic
  resync when client and server drift. With HTMX you rebuild this yourself
  the day you need "server state changed, update every open tab".
- **Client-side logic without JavaScript.** Event delegation, DOM queries,
  the router, HTTP calls, all are Go APIs running in wasm. Logic too chatty
  for a server round trip (input filtering, drag state, canvas work) stays
  in Go instead of becoming the JS island every hypermedia app eventually
  grows.
- **One artifact.** No npm, no bundler, no frontend lockfile. `rfw build`
  produces the wasm client and static assets; the host is a normal Go
  binary. Your CI is `go test ./...` and a build.

## Where rfw loses today

Read this section as carefully as the previous one.

- **Wasm binary size.** A Go wasm binary is measured in megabytes, not
  kilobytes. rfw serves Brotli-compressed `.wasm.br` and the runtime
  hydrates server-rendered HTML, which softens first paint, but the
  download and instantiation cost is real and HTMX/Datastar simply do not
  pay it. On slow networks this is a genuine disadvantage.
- **Ecosystem maturity.** rfw is pre-1.0. APIs still move between releases,
  the component ecosystem is small, documentation has gaps, and you will
  sometimes read framework source to answer a question. templ, HTMX, and
  the React world have years of Stack Overflow behind them; rfw does not.
- **Hiring and onboarding.** Nobody you interview will know rfw. Go
  developers pick it up quickly, but you are training every hire, and the
  skills are less transferable than React or plain hypermedia patterns.
  For some teams that is disqualifying, and that is a fair call.
- **Operational model.** SSC means persistent WebSockets and per-session
  server state, which makes horizontal scaling and failover more involved
  than stateless fragment rendering. Debugging inside wasm is also rougher
  than debugging JavaScript, browser devtools see the wasm frame, not your
  Go source.

## Choosing

If your app is pages, forms, and the occasional dynamic fragment, use
templ + HTMX and be happy; rfw would be more machinery than the problem
needs. If you need server-pushed updates but want to stay minimal, Datastar
is a credible middle ground. Reach for rfw when the app is state: real-time
dashboards, internal tools, control planes, anything where server state
must be reflected in every connected client instantly and the client itself
has logic worth type-checking.
