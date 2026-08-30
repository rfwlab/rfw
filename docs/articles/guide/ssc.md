# Server Side Computed (SSC)

SSC splits a component in two: an HTML component that renders in the
browser, and a host component that runs in your Go server process. A
persistent WebSocket at `/ws` keeps them synchronized. The browser loads a
lightweight wasm binary that hydrates the HTML; your logic, queries and
secrets stay on the server. For the trust boundary and hardening options,
read the [SSC security model](ssc-security.md).

## Host components

A host component is a named handler registered in the server binary:

```go
host.Register(host.NewHostComponentWithSession("Counter",
    func(session *host.Session, payload map[string]any) any {
        store := session.StoreManager().NewStore("counter")
        if inc, ok := payload["increment"].(bool); ok && inc {
            current, _ := store.Get("value").(int)
            store.Set("value", current+1)
        }
        return map[string]any{"value": store.Get("value")}
    }))
```

Whatever the handler returns is JSON-marshaled and sent to the client,
which applies each key to the matching `data-host-var` element and host
signal. Each connection gets a `*host.Session` with an isolated store
manager and a context bag (`ContextGet`/`ContextSet`); keep per-user data
there, never in global stores.

## Binding from the client

Templates reference host variables with `{h:name}` and host commands with
`@h:name`:

```html
<root>
  <p>{h:value}</p>
  <button @h:increment>+1</button>
</root>
```

With the composition API, host signal types (`t.HInt`, `t.HString`, ...)
declare server-synced bindings on the struct; every host field name is
registered against the server-side component of the same name. HTML
components link to their host explicitly with `AddHostComponent(name)`.

A declared host component is bound once per mounted lifecycle, not once per
render: a component that is rendered but never mounted binds nothing, since no
unmount would come to release it. Unmounting the component, which is what
leaving a route does, drops the binding, discards its queued reconnect
initialization and unsubscribes from the host, so the feed stops and a late push
reaches neither the old component root nor its host signals: a frame that was
already in flight when the cleanup ran is discarded rather than applied, and
delivery only ever addresses the exact root the binding owns, whatever
characters its id carries. A host signal setter may unmount or remount its own
component, which releases and registers the binding from inside the frame being
applied; that ends the frame and never deadlocks. Mounting again binds it
again.

Code that drives the runtime directly keeps both entry points.
`hostclient.RegisterComponent(id, name, vars)` binds until another registration
replaces it, and `hostclient.RegisterComponentOwned` returns the idempotent
cleanup for a component that can unmount. While the client is disconnected the
queue holds only the latest state per host component, so entering and leaving a
route offline does not pile up reconnect messages.

On the wire, `hostclient.Send(name, payload)` delivers a payload to the
host component. Repeated identical messages are delivered as-is; call
`hostclient.EnableSendDedup(name)` if a channel should drop identical
payloads sent within a 5 second window.

## Typed actions and forms

Typed actions reject unknown request fields before the handler runs and return
a correlated response:

```go
type RenameRequest struct {
    UserID string `json:"userId"`
    Name   string `json:"name"`
}

type RenameResponse struct {
    Updated bool `json:"updated"`
}

err := host.RegisterAction("users.rename",
    func(ctx context.Context, session *host.Session, request RenameRequest) (RenameResponse, error) {
        return renameUser(ctx, session, request)
    },
    host.WithActionAuthorizer(func(ctx context.Context, session *host.Session, request RenameRequest) error {
        return authorizeUserEdit(ctx, session, request.UserID)
    }),
)
```

The wasm client uses the same request and response types:

```go
result, err := hostclient.Call[RenameRequest, RenameResponse](
    ctx,
    "users.rename",
    RenameRequest{UserID: "42", Name: "Ada"},
)
```

`host.RegisterForm` adds field validation before submission. Invalid values
return `host.FormResponse` with `Valid == false` and a `Fields` map. The client
counterpart is `hostclient.SubmitForm`.

Use typed actions for commands that need strict input, authorization, a
deadline, and a result. Existing host components remain useful for continuous
host-variable synchronization and event streams.

## Pushing from the server

`host.Broadcast(name, payload)` sends to every connection subscribed to a
component; scope per-user data with
`host.Broadcast(name, payload, host.WithSessionTarget(sessionID))`.

A host component can also register an init snapshot, a rendered HTML
fragment the client injects wholesale on resync. Build snapshots with the
escaping helpers (`host.Span`, `host.Div`, `host.P`, `host.Tag`), which
HTML-escape values by default; `host.RawTag` and `host.Raw` are the
explicit trust APIs for markup you generated yourself.

## Serving

`host.StartAuto()` (or `host.Start(root)`) serves the client build over
HTTP and HTTPS and registers the `/ws` endpoint. For more control, build
the mux yourself:

```go
mux := host.NewMux(root,
    host.WithOriginAllowlist("https://app.example.com"),
    host.WithAuthFunc(func(r *http.Request) bool { return validCookie(r) }),
    host.WithSSCSessionInitializer(func(r *http.Request, session *host.Session) error {
        session.ContextSet("user", userFromRequest(r))
        return nil
    }),
    host.WithSSCAuthorizer(func(ctx context.Context, session *host.Session, message host.Inbound) error {
        return authorizeMessage(ctx, session, message)
    }),
)
host.ListenAndServeWithMux(":8080", mux)
```

`ssc.NewSSCServer(addr, root, opts...)` accepts the same guard options and
adds an event bus (`ssc.SubscribeSSC`) fed by every inbound message.

`host.WithSSCLimits` overrides frame size, connection count, per-session
message rate, handler and write deadlines, the bounded outbound queue, resume
lifetime, and replay history. The defaults are:

```go
host.SSCLimits{
    MaxMessageBytes:   1 << 20,
    MaxConnections:    4096,
    MaxSessions:       8192,
    MessagesPerMinute: 600,
    HandlerTimeout:    15 * time.Second,
    WriteTimeout:      10 * time.Second,
    OutboundQueueSize: 64,
    ResumeTTL:         2 * time.Minute,
    ReplayMessages:    256,
}
```

The protocol assigns sequence numbers in both directions and acknowledges
received messages. The client retains unacknowledged writes. On reconnect it
presents an opaque resume token, the host reattaches the session, and retained
host responses are replayed. Broadcasts enqueue independently per connection,
so one slow client cannot delay the others. A connection that exhausts its
outbound queue is sent `resync_required` when possible and disconnected rather
than silently losing a sequenced message. Keep the queue smaller than replay
history so overflowed delivery remains resumable. `hostclient.ConnectionStateSignal`
reports `connecting`, `connected`, `disconnected`, or `desynced`.
Use `host.WithoutSSCResume()` when detached session state must be discarded
immediately.

A resume token reattaches whatever session it names, so an authenticated
deployment decides who may present it with
`host.WithSSCResumeAuthorizer(func(r *http.Request, session *host.Session) error)`.
The callback runs before the session is touched and a refusal serves the caller
a new session instead, leaving the retained one available to its owner. See
[SSC security](ssc-security.md).

During development `rfw dev` detects `"type": "ssc"` in `rfw.json`, builds
the host binary and restarts it on every rebuild.
