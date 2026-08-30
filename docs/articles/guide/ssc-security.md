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
{
  "component": "HelloHost",
  "payload": { "cmd": "increment" },
  "sequence": 4,
  "ack": 7,
  "resumeToken": "<opaque token>"
}
```

The client sends: an `{"init": true}` payload for each bound component on
connect, whatever payloads your client code passes to `hostclient.Send`,
and automatic `resync` requests when hydration detects a mismatch.

**Server to client** (`host.Outbound`):

```json
{
  "component": "HelloHost",
  "payload": { "greeting": "hello" },
  "session": "<id>",
  "sequence": 8,
  "ack": 4,
  "resumeToken": "<opaque token>"
}
```

The payload is one of:

- **Host variable values.** Whatever your handler returns is
  `json.Marshal`ed and sent as-is. The client applies each key to the
  matching `data-host-var` element (as text) and to the matching host
  signal (`t.HInt`, `t.HString`, ...).
- **An init snapshot** (`host.InitSnapshot`): a rendered HTML fragment plus
  a list of variable names, sent in response to a `resync` request. The
  client injects `snapshot.HTML` into the component root wholesale.
- **Delivery metadata.** Session ID, resume token, sequence, and
  acknowledgement fields support ordered delivery and reconnect replay.

Typed actions add an action name, request ID, result, or public
`host.ActionError`. The wire does not carry code, hidden store contents, or
state that a handler did not explicitly return.

## What stays on the server

- **Your Go code.** Host handlers are never serialized; only their return
  values are. Business logic, queries, and validation are invisible to the
  client.
- **Secrets and handles.** Database connections, API keys, and anything
  else living in the host process never crosses the wire unless a handler
  puts it in a return value. The corollary: *everything a handler returns
  becomes public*. Do not return raw database rows or internal structs;
  build the payload explicitly.
- **Session state.** Each new WebSocket identity gets a `host.Session` with
  an isolated `state.StoreManager` and a context bag
  (`ContextGet`/`ContextSet`). None of it is sent to the client except the
  random session ID and resume token. Detached state remains available for
  the configured resume lifetime. Note that `state.GlobalStoreManager` is
  shared across all sessions; keep per-user data in the session's store
  manager, never in global stores.

One caveat about the client side of the boundary: the wasm binary ships to
the browser. Any string compiled into client code (tokens, endpoints,
"hidden" logic) is extractable. Treat client Go code exactly like you would
treat JavaScript.

## Authentication and authorization

This is the most important paragraph on this page. As implemented today:

- By default the `/ws` endpoint accepts any connection. `wsHandler`
  allocates a session after the upgrade guards pass. There is no login or
  `Origin` check unless you enable the guard options below.
- Legacy host components are addressed by name in a global registry. Any
  connected client can send any payload to any registered component. The
  component name in a message is data chosen by the client, not a routing
  decision you made.
- The session ID and resume token identify transport state; neither proves a
  user identity. Both are generated with `crypto/rand`, but they belong to
  whoever established the connection.

Apply these controls together:

1. **Gate the upgrade with the built-in guards.** `host.NewMux` and
   `ssc.NewSSCServer` accept `host.WithOriginAllowlist(...)` and
   `host.WithAuthFunc(...)`:

   ```go
   mux := host.NewMux(root,
       host.WithOriginAllowlist("https://app.example.com"),
       host.WithAuthFunc(func(r *http.Request) bool {
           return validSessionCookie(r)
       }),
   )
   ```

   The allowlist rejects upgrades whose `Origin` header is missing or
   unlisted (403) before a session is allocated; the auth func sees the
   full upgrade request (cookies, headers) and returning false rejects
   with 401. Both default to off. Your own middleware in front of the mux
   (or a reverse proxy enforcing auth and an `Origin` check) works too.
2. **Bind identity to the session.** Use
   `host.WithSSCSessionInitializer` to copy verified request identity into the
   session context before its first message:

   ```go
   host.WithSSCSessionInitializer(func(r *http.Request, session *host.Session) error {
       session.ContextSet("user", authenticatedUser(r))
       return nil
   })
   ```

   The initializer runs for new sessions only. A resumed session keeps the
   identity it was created with, which is the point of the next control.
3. **Authorize resume.** A resume token reattaches the session it names. The
   upgrade guard proves the new request is authenticated; it does not prove it
   is authenticated *as the user the token belongs to*. Without
   `host.WithSSCResumeAuthorizer`, any client that obtains a valid token
   reattaches that session, private store contents included:

   ```go
   host.WithSSCResumeAuthorizer(func(r *http.Request, session *host.Session) error {
       user, ok := session.ContextGet("user")
       if !ok || user != authenticatedUser(r) || revoked(r) {
           return errors.New("resume denied")
       }
       return nil
   })
   ```

   The callback runs before the session is attached, its retention timer
   cleared or its connection replaced. Refusing changes nothing: the caller is
   served a new session and receives the `resume_rejected` control frame that
   clears its stale token, while the retained session stays detached and
   resumable by its owner until the TTL expires. Two upgrades racing on the
   same token cannot cross-attach; the loser gets a new session too. An
   authenticated multi-user deployment must configure a resume authorizer or
   disable resume with `host.WithoutSSCResume()`. The default remains
   token-only for compatibility with single-user and unauthenticated
   deployments.
4. **Authorize messages and actions.** `host.WithSSCAuthorizer` can reject
   every decoded message before dispatch. `host.WithActionAuthorizer` applies
   a typed policy after an action request is decoded. Legacy component
   handlers should still perform their own object-level checks.

## Threat notes

- **All client input is untrusted.** Legacy handler payloads are
  `map[string]any` decoded from client JSON. The client is free to send
  payloads your UI would never produce: unexpected keys, wrong types,
  hostile values, messages for components the user never rendered.
  Validate types and ranges in the handler, on the server, every time. Typed
  actions use `json.Decoder.DisallowUnknownFields`, but tags and field types
  are only structural validation; validate ranges and business rules too.
  Client-side validation in wasm is UX, not security.
- **Escape what you render.** The server-side HTML helpers (`host.Span`,
  `host.Div`, `host.P`, `host.Tag`) HTML-escape the value by default, so
  user-derived data in a host variable's initial render stays text.
  `InitSnapshot.HTML` is still injected into the DOM as raw HTML on the
  client: build snapshots from the escaping helpers, and reach for
  `host.RawTag` (unescaped value) or `host.Raw` (unescaped fragment) only
  with markup you generated or sanitized yourself. Subsequent
  host-variable *updates* are applied as text content and are safe from
  injection.
- **Broadcast scope.** `host.Broadcast(name, payload)` sends to every
  connection subscribed to that component. Use
  `host.WithSessionTarget(sessionID)` for per-user data; broadcasting a
  payload that contains one user's data sends it to all users on that
  component.
- **Client-side delivery scope.** A payload is applied only to the root the
  binding names (`[data-component-id]`) and to that component's host signals,
  and only while that binding is registered. A frame that arrives after its
  component unmounted, including one already in flight when the route changed,
  is dropped rather than applied to the page shell, to whatever mounted next, or
  to a signal the unmounted component still holds. The id is escaped before it
  reaches a selector, so an id carrying CSS metacharacters cannot widen the
  match to a root the binding does not own. That bounds the damage a stale
  broadcast can do to the DOM; it is not an authorization control, since the
  browser still received the payload. Scope the data itself with
  `host.WithSessionTarget`.
- **Transport security.** A resume token can reattach detached session state.
  Treat it as a bearer credential, do not log it, use `wss://`, and pair it
  with `host.WithSSCResumeAuthorizer` so a stolen token is not enough on its
  own.
  `host.Start` serves HTTP and, on the next port,
  HTTPS with a self-signed certificate generated at boot. That certificate
  is a development convenience. In production, terminate TLS with real
  certificates (typically at a reverse proxy) so the WebSocket runs over
  `wss://`; otherwise session IDs and payloads travel in cleartext.
- **Resource exhaustion.** The default endpoint caps frames at 1 MiB,
  connections at 4096, retained sessions at 8192, messages at 600 per session
  per minute, typed action execution at 15 seconds, writes at 10 seconds, the
  per-connection outbound queue at 64 pending batches, and replay history at 256
  messages. A full outbound queue disconnects the client for a clean resume
  instead of dropping a sequenced frame. Override these
  values with `host.WithSSCLimits` for the deployment. Keep proxy limits too.
  Action handlers should observe their context; a handler that ignores
  cancellation may continue running after the timeout response.

rfw keeps application code and private state on the server, but it does not
ship an identity provider. Treat every inbound message as hostile until the
upgrade guard, session initializer, and authorization policy have established
the caller and permitted the operation.
