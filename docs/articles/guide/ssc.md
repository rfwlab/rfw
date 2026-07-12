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

On the wire, `hostclient.Send(name, payload)` delivers a payload to the
host component. Repeated identical messages are delivered as-is; call
`hostclient.EnableSendDedup(name)` if a channel should drop identical
payloads sent within a 5 second window.

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
)
host.ListenAndServeWithMux(":8080", mux)
```

`ssc.NewSSCServer(addr, root, opts...)` accepts the same guard options and
adds an event bus (`ssc.SubscribeSSC`) fed by every inbound message.

During development `rfw dev` detects `"type": "ssc"` in `rfw.json`, builds
the host binary and restarts it on every rebuild.
