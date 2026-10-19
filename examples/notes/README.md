# Notes over HTTP

A complete create/read/update/delete example using `http.Request`, runtime
escaped HTML, delegated element handlers and mount-scoped response guards.
The native Go server stores notes in memory. A browser reload keeps its notes;
stopping that server loses them. No database, authentication or shared-user
security is provided. This is an example, not a public notes service.

## Build and run from this checkout

Install the checkout's CLI, not an unrelated released version:

```sh
# At the framework repository root.
go install ./cmd/rfw
cd examples/notes
"$(go env GOPATH)/bin/rfw" build
go run ./server --port 8080 --assets build/client
```

Open http://127.0.0.1:8080. The server binds only that loopback address and
refuses different Host/Origin values. Do not expose it through a proxy or bind
it to a public interface. The static bundle does not need SSC or WebSocket.
Unlike the purely static examples, `rfw dev` alone cannot supply this API.
Rebuild the client after code changes; restarting the API resets its notes.

## Contract

- GET `/api/notes` returns the current array.
- POST `/api/notes` accepts `{"text":"A note"}`.
- PUT `/api/notes/1` accepts the same body and requires `If-Match: 1` for the
  version previously read. DELETE uses the same identity/version precondition.
- Successful writes return the complete current array. IDs are stable for the
  API process lifetime, never row indexes. A stale update or delete returns409.
- Text is limited to200 bytes, bodies to4096 bytes and retained notes to100.
  The browser's maxlength is only a convenience; the server enforces byte limits.
  API responses are not cached.

The client disables its form during a request and keeps the input after an
error. A transport failure has unknown mutation outcome: refresh the list and
inspect it before deciding to retry. It never automatically retries a write.
An explicit Refresh does not silently replace the editor's stale version; select
Edit on the current row to load its text/version after reviewing a conflict.

Leaving for About closes the component scope. The callback captures that mount's
context and ignores a late completion. `http.Request` currently has no cancellation
handle, so ignoring the response does not abort fetch or undo a server write.
No callback touches a replacement component's DOM.

## Checks

```sh
go test -race ./server
GOOS=js GOARCH=wasm go build ./...
```

The native tests cover CRUD, stable IDs, stale edit/delete, uncached reads and
invalid/oversized bodies. Browser acceptance must drive Save/Edit/Delete and
reload against the real server. See the [tutorial](../../docs/articles/guide/http-crud-tutorial.md)
for the implementation path and the production work this example deliberately
leaves to the application.
