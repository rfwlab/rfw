# Build a CRUD page against a Go HTTP API

Use [examples/notes](../../../examples/notes/) as the runnable source. It adds
an API to the existing dynamic-list pattern rather than introducing another
state library or a JavaScript build pipeline. Use the framework checkout's CLI
so the compiler, loader and local `replace` dependency refer to the same source.

```sh
# Run at the framework repository root with the patched Go from go.mod.
go install ./cmd/rfw
cd examples/notes
"$(go env GOPATH)/bin/rfw" build
go run ./server --port 8080 --assets build/client
```

Open http://127.0.0.1:8080. Add text, select Save note, edit a row, save it and
reload the browser. Delete a row. The browser reads the same real API after
reload. Stop the server only when you are ready to discard its in-memory notes.
The server is a loopback teaching fixture, not an authenticated deployment.

## 1. Share the response model

`model/note.go` defines ID, text and version. The client and native server import
this same Go type. ID identifies the entity, not its position in the displayed
slice. Version identifies the state the user reviewed. A new note receives a new
ID even when an older note was deleted.

`server/main.go` owns a mutex-protected slice and exposes GET/POST/PUT/DELETE.
It validates text/body/count limits and unknown input fields. Update and delete
compare `If-Match` with the current version while holding the same mutex as the
write. A mismatch returns409 without changing the note. A production application
must implement these checks in its own transaction, not use a process mutex as
a distributed database lock.

## 2. Mount the page and delegate its actions

`pages/index.go` returns a fresh NotesComponent. The pages build plugin registers
it and `pages/about.go` as real routes. `notes_component.rtml` contains a labelled
input, a submit handler, an aria-live status and the list container. The form's
submit handler calls `PreventDefault`, so Save works with a button or keyboard.

The component registers Save, Edit, Delete, Refresh and New note once. Its
runtime row builder escapes text with `html.EscapeString`, attaches stable IDs
and versions as data attributes and passes only the generated markup through
`dom.ExpandEvents`. Event expansion is not HTML sanitization. Do not substitute
untrusted text into event handler names, selector syntax or raw markup.

## 3. Treat every HTTP outcome explicitly

`http.Request` accepts RequestOptions and completes with status/body. The example
uses uncached requests for reads as well as writes, then decodes the200 response
with `encoding/json`. It displays409 as a review conflict, other non-200 statuses
as refusal, and invalid JSON as a response error. Status0 means a browser,
transport or body-read failure. A failure callback can run before Request returns;
set busy state before starting it.

The callback only updates the list and clears the submitted form after confirmed
success. Draft text survives a refusal or transport failure. There is no automatic
retry: a failed response does not prove the server rejected a mutation. Refresh
first and inspect the current list. After a conflict, explicitly select Edit on
the fresh row before submitting that version. A second browser tab can demonstrate
this: both load the same note, one saves, and the other's older edit returns409.

## 4. Keep responses within their mount

Before Request, the component captures `c.Scope().Context()`. The callback checks
that exact context before touching busy state or DOM. Navigating to About closes
the scope; a pending response cannot overwrite a later mount's list or status.
Do not ask `c.Scope()` for a new context inside a late callback: the same component
may have acquired a different scope by then.

This is a stale-response guard. Request does not currently expose an abort handle.
Unmount does not cancel the HTTP operation or reverse a write already accepted
by the server. Stronger application requirements need request cancellation,
reconciliation and durable idempotency on the server side.

## 5. Verify behavior, then choose production contracts

Run `go test -race ./server`, build with `rfw build`, and exercise the actual UI.
Check browser reload, empty rows, HTML-looking text, keyboard submission, stale
edits in two tabs, request refusal, draft retention and navigation during a delayed
response. Browser interception can inject a response delay or transport failure;
it must not replace the successful CRUD API with fixtures and call that end to end.

The example binds127.0.0.1 and checks Host/Origin. It has no accounts, permissions,
CSRF/session design, durable database, audit/outbox, pagination or process-restart
recovery. Select those contracts before sharing an application with users. The
framework supplies rendering, routing, lifecycle and HTTP primitives; your backend
still owns authorization, validation, transactions and persistence.
