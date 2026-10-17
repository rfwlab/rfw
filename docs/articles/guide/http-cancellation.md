# Browser HTTP cancellation

`http.RequestContext` and `http.RequestBytesContext` are unreleased browser
APIs for Go `js/wasm` clients. They accept `context.Context` before the existing
URL, options and callback arguments. `RequestOptions` is unchanged, including
browser-side `BodyValue`. The existing `Request` and `RequestBytes` keep their
unbounded lifetime. Cached `FetchJSON`/`FetchText` and native HTTP configuration
are separate APIs; this change does not add deadlines to those calls.

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
http.RequestContext(ctx, "/api/notes", http.RequestOptions{
    Method: "POST",
    Headers: map[string]string{"Content-Type": "application/json"},
    Body: `{"text":"Review release"}`,
}, func(status int, body string) {
    defer cancel()
    // status0 or an ambiguous server error needs explicit state recovery.
    // Do not replay an unknown write automatically.
})
```

Keep the cancel function with the logical owner of a request and call it when
that owner changes. A physical component repaint need not change ownership.
Cancel in the completion callback too, to release the context's timer. Imports
are the standard `context`/`time` packages and `github.com/rfwlab/rfw/v2/http`.

## Completion and resource contract

- A context cancelled before the call completes synchronously with status0
  without reaching fetch. A nil context behaves like `context.Background`.
- Native browser AbortController binds cancellation to fetch and its response
  body. Pending text and binary reads are aborted too. Non-cancellable contexts
  do not allocate a controller. A missing controller fails closed with status0
  for a cancellable context.
- HTTP responses keep their status, including4xx/5xx. Network, body-read and
  cancellation failures deliver status0 plus empty text or nil bytes. Inspect
  the supplied context's error if the caller needs its cancellation reason.
- One promise chain reaches either the body result or rejection. Callback
  functions are released after the chain settles; the context listener is
  stopped before the caller callback. Cancellation does not release functions
  while a native browser promise could still invoke them.
- There is no retry, cache or default timeout. The caller chooses its deadline.
  Callbacks run through the existing browser callback guard; caller panics stay
  caller errors and are not converted into a transport failure.

This contract requires native browser fetch/AbortController behavior. It cannot
force a replacement fetch implementation that ignores its signal to settle.
It does not prove that a remote operation stopped: a server may already have
accepted a write before the response was aborted. Such writes need an explicit
read, receipt or idempotency protocol before a user chooses another action.

## Verification

`http/request_context_test.go` covers precancellation, pending fetch deadlines,
body cancellation, one completion, missing-controller refusal and cancellation
listener removal after success. Existing request lifecycle/status/binary/form
and caller-panic regressions apply to the shared implementation.

The maintained `bench/http-context.mjs` builds `bench/http-context` and starts an
owned loopback HTTP server. It uses the existing bench Playwright dependency:

```sh
RFW_HTTP_RESULTS=/absolute/owned/output node bench/http-context.mjs
```

The output directory must be writable and empty of conflicting files. Set
`GO_BINARY`, `CHROME_PATH`, `PLAYWRIGHT_BROWSERS_PATH` or `RFW_ROOT` when needed.
No dependency installation is performed by the harness. Chromium, Firefox and
WebKit check real pending headers, a pending binary body, precancellation,
repeated aborts, exact binary bytes, legacy404 behavior and a server counter
which remains committed after an aborted response. A response method wrapper
records when native arrayBuffer reading starts without replacing its result.
Each pending transport must close on the server. Records and screenshots are
saved with results. This is a primitive contract test, not product UX, production
load, proxy, mobile-device, byte-budget or all-application cancellation proof.
