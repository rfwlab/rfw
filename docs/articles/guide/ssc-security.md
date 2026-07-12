# SSC security model

Server Side Computed (SSC) components split your app across a trust
boundary: host components run in your Go server process, the wasm client
runs in the user's browser, and a WebSocket connects them. This page
documents exactly what crosses that boundary, what never does, and which
security responsibilities remain yours. Everything here is grounded in the
implementation (`host/`, `hostclient/`, `ssc/`).

## What travels over the wire

All SSC traffic flows over a single WebSocket, served at `/ws` by
`host.NewMux` (or `ssc.NewSSCServer`). Messages are JSON.

**Client to server** (`host.Inbound`):

```json
{ "component": "HelloHost", "payload": { "cmd": "increment" } }
```

The client sends: an `{"init": true}` payload for each bound component on
connect, whatever payloads your client code passes to `hostclient.Send`,
and automatic `resync` requests when hydration detects a mismatch.

**Server to client** (`host.Outbound`):

```json
{ "component": "HelloHost", "payload": { "greeting": "hello" }, "session": "<id>" }
```

The payload is one of:

- **Host variable values.** Whatever your handler returns is
  `json.Marshal`ed and sent as-is. The client applies each key to the
  matching `data-host-var` element (as text) and to the matching host
  signal (`t.HInt`, `t.HString`, ...).
- **An init snapshot** (`host.InitSnapshot`): a rendered HTML fragment plus
  a list of variable names, sent in response to a `resync` request. The
  client injects `snapshot.HTML` into the component root wholesale.
- **The session ID**, echoed on every outbound message and sent on init.

So the wire carries rendered fragments, plain variable values, and session
IDs. It does not carry code, diffs of your server state, or store contents
you did not explicitly return.

## What stays on the server

- **Your Go code.** Host handlers are never serialized; only their return
  values are. Business logic, queries, and validation are invisible to the
  client.
- **Secrets and handles.** Database connections, API keys, and anything
  else living in the host process never crosses the wire unless a handler
  puts it in a return value. The corollary: *everything a handler returns
  becomes public*. Do not return raw database rows or internal structs;
  build the payload explicitly.
- **Session state.** Each WebSocket connection gets a `host.Session` with
  an isolated `state.StoreManager` and a context bag
  (`ContextGet`/`ContextSet`). None of it is sent to the client except the
  random session ID. Note that `state.GlobalStoreManager` on the server is
  shared across all sessions; keep per-user data in the session's store
  manager, never in global stores.

One caveat about the client side of the boundary: the wasm binary ships to
the browser. Any string compiled into client code (tokens, endpoints,
"hidden" logic) is extractable. Treat client Go code exactly like you would
treat JavaScript.

## Authentication and authorization: rfw does neither

This is the most important paragraph on this page. As implemented today:

- The `/ws` endpoint accepts any connection. `wsHandler` allocates a
  session for every socket, unconditionally. There is no login, no token
  check, and no `Origin` allowlist in the framework.
- Host components are addressed by name in a global registry. Any connected
  client can send any payload to any registered component. The component
  name in a message is data chosen by the client, not a routing decision
  you made.
- The session ID identifies a connection; it does not prove an identity.
  It is 16 bytes from `crypto/rand`, which makes it unguessable, but a
  session is created for whoever connects.

The recommended pattern until you have something better:

1. **Authenticate at the HTTP layer.** Build your mux with `host.NewMux`
   and wrap it (or serve it via `host.ListenAndServeWithMux`) behind your
   own middleware that validates a session cookie or token *before* the
   WebSocket upgrade, and rejects unauthenticated upgrades of `/ws`.
   A reverse proxy enforcing auth and an `Origin` check works too.
2. **Bind identity to the session.** After verifying credentials (from the
   upgrade request, or from a first authenticated message), store the user
   identity with `session.ContextSet("user", ...)`. Handlers receive the
   `*host.Session` and can read it back.
3. **Authorize in every handler.** Each `Serve`/handler call must check
   that the session's identity is allowed to perform the requested action.
   There is no framework hook that does this for you.

## Threat notes

- **All client input is untrusted.** Handler payloads are
  `map[string]any` decoded from client JSON. The client is free to send
  payloads your UI would never produce: unexpected keys, wrong types,
  hostile values, messages for components the user never rendered.
  Validate types and ranges in the handler, on the server, every time.
  Client-side validation in wasm is UX, not security.
- **Escape what you render.** The server-side HTML helpers (`host.Span`,
  `host.Div`, `host.P`, `host.Tag`) interpolate the value into markup
  without HTML-escaping it, and `InitSnapshot.HTML` is injected into the
  DOM as raw HTML on the client. If any user-derived data ends up in a
  host variable's initial render or in a snapshot, HTML-escape it yourself
  (e.g. `html.EscapeString`) before building the fragment. Subsequent
  host-variable *updates* are applied as text content and are safe from
  injection.
- **Broadcast scope.** `host.Broadcast(name, payload)` sends to every
  connection subscribed to that component. Use
  `host.WithSessionTarget(sessionID)` for per-user data; broadcasting a
  payload that contains one user's data sends it to all users on that
  component.
- **Transport security.** `host.Start` serves HTTP and, on the next port,
  HTTPS with a self-signed certificate generated at boot. That certificate
  is a development convenience. In production, terminate TLS with real
  certificates (typically at a reverse proxy) so the WebSocket runs over
  `wss://`; otherwise session IDs and payloads travel in cleartext.
- **Resource exhaustion.** Nothing in the framework rate-limits messages
  or connections. A hostile client can open sockets and spam handlers.
  Apply connection and message limits at your proxy, and keep handlers
  cheap or queue their work.

The summary: rfw's transport keeps your code and state on the server by
construction, but it deliberately ships no identity layer. Assume every
inbound message is from an anonymous, possibly hostile client until your
own middleware and handlers have proven otherwise.
