# Server-Side Computed (SSC)

**SSC** runs application logic on the **server**, while the browser loads a lightweight **Wasm** bundle to hydrate server-rendered HTML. Server and client keep state synchronized over a persistent **WebSocket**. Server-originated bindings and commands are prefixed with `h:`.

**SSC is required in v2**, `rfw init` scaffolds with SSC enabled by default.

---

## What is SSC?

In a standard client-side app, components render entirely in the browser. With SSC, the **host** process (a Go server) renders HTML and executes privileged logic. The browser receives fully rendered markup, then the Wasm bundle **hydrates** it, attaching event handlers and reactive bindings, so the page becomes interactive. Updates involving server state travel over the WebSocket.

**Key idea:** heavy and sensitive work stays on the server; the client stays tiny and reactive.

---

## Why SSC?

- **Faster time-to-content**, users see meaningful HTML before Wasm finishes loading
- **Unified Go model**, write server and client logic in Go; share types and business rules
- **Better SEO**, crawlers index server-rendered HTML
- **Tighter security**, secrets and privileged actions remain on the server

---

## Trade-offs

- **Two artifacts**, a client Wasm bundle and a **host** binary
- **Server resources**, rendering and state sync consume CPU/memory; plan caching and capacity
- **Environment constraints**, browser-only code must run client-side; server code cannot use browser globals
- **Hydration care**, server HTML and client expectations must match to avoid mismatches

---

## Enabling SSC

SSC is the default build type. In `rfw.json`:

```json
{
  "build": { "type": "ssc" }
}
```

Running `rfw build` produces:

- `build/client/`, Wasm bundle and client assets
- `build/host/`, the host server binary

---

## Host/Client Architecture

### Client component uses host signal types

Use `t.HInt`, `t.HString`, `t.HBool`, `t.HFloat` fields for host-synced values:

```go
//go:build js && wasm

package components

import (
    "embed"

    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/types"
)

//go:embed GreetingComponent.rtml
var fs embed.FS

func init() {
    composition.RegisterFS(&fs)
}

type GreetingComponent struct {
    composition.Component

    ClientMsg types.String
    Visit     types.HInt
}
```

`t.HInt` (and other host types) are both signals and host component declarations. `composition.New` auto-calls `AddHostComponent` for each host-type field.

### Struct-based host component

Define a struct implementing the `host.Component` interface:

```go
package main

import (
    "github.com/rfwlab/rfw/v2/host"
    "github.com/rfwlab/rfw/v2/ssc"
)

type VisitHandler struct{}

func (v *VisitHandler) Name() string { return "Visit" }

func (v *VisitHandler) Serve(h host.Handler) {
    h.On("UpdateVisits", func(p host.Payload) host.Payload {
        visits++
        return host.Payload{"visit": visits}
    })
}

func main() {
    host.RegisterComponent(&VisitHandler{})

    sscSrv := ssc.NewSSCServer(":8080", "client")
    sscSrv.ListenAndServe()
}
```

### Function-based host (simple cases)

```go
host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
    return map[string]any{"hostMsg": "hello from server"}
}))
```

### Template mixing client and host values

```rtml
<root>
  <p>Client: {@expr:clientMsg}</p>
  <p>Host: {h:hostMsg}</p>
  <button @on:click:h:updateTime>refresh</button>
</root>
```

- `@expr:clientMsg` reads a client signal
- `{h:hostMsg}` is pushed by the host over WebSocket
- `@on:click:h:updateTime` invokes a **host** command

---

## HTML Helpers for Host Rendering

The `host` package provides HTML builders that auto-add `data-host-var` and `data-host-expected` attributes:

```go
import "github.com/rfwlab/rfw/v2/host"

body := host.Div(
    host.Span("Visits: ", host.Var("visit"), host.Expected("visit")),
)
```

| Helper | Description |
| --- | --- |
| `host.Span(children...)` | `<span>` with optional bindings |
| `host.Div(children...)` | `<div>` container |
| `host.P(children...)` | `<p>` paragraph |
| `host.Tag(name, children...)` | Generic element |
| `host.Text(s)` | Text node |
| `host.Raw(html)` | Raw HTML string |
| `host.Join(nodes...)` | Concatenate nodes |
| `host.Var(name)` | `data-host-var` binding marker |
| `host.Expected(name)` | `data-host-expected` binding marker |

---

## Development Workflow

`rfw dev` rebuilds on changes, serves assets, and, when a `host/` directory is present, builds and runs the host binary from `build/host/host` so you can iterate locally.

Use `--debug` to enable profiling endpoints (`/debug/vars`, `/debug/pprof/`).

---

## Hydration

1. The host responds with **fully rendered HTML**
2. The browser downloads the Brotli-compressed `app.wasm.br` (falling back to `app.wasm`) and initializes rfw
3. rfw **hydrates** the server markup, attaches event handlers and reactive bindings
4. A WebSocket connects; `h:` values and commands synchronize with the host

**Avoid mismatches:** ensure server-rendered data is consistent with what the client expects on first paint. Avoid random values and timezone-dependent formatting during initial render. If divergence is unavoidable, render those parts client-only after mount.

---

## Session-Aware Host Components

For per-request logic, use `NewHostComponentWithSession` or the `OnSession` handler method:

```go
host.Register(host.NewHostComponentWithSession("UserPanel", func(session *host.Session, payload map[string]any) any {
    userID := session.ID()
    return map[string]any{"userId": userID}
}))
```

Session-specific stores are available via `hc.StoreManager(session)`.

---

## Broadcasting

Push updates from the server to connected clients:

```go
ssc.Broadcast("GreetingHost", map[string]any{"hostMsg": "updated!"})

// Target a specific session
ssc.Broadcast("GreetingHost", map[string]any{"hostMsg": "private"}, ssc.WithSessionTarget(sessionID))
```

---

## Code Structure

Keep shared types and logic in regular Go packages imported by both sides:

- **Shared:** DTOs, validators, domain logic
- **Host only:** database access, secrets, privileged integrations
- **Client only:** DOM-dependent code, animations, browser APIs

---

## Caching & Performance

- Cache host-generated responses or fragments
- Coalesce frequent server updates and send compact diffs through `h:` bindings
- Profile with `--debug` during development

---

## Deployment

- Deploy the **host** binary behind a reverse proxy; serve the **client** bundle from `build/client/`
- Ensure the WebSocket endpoint (`/ws`) is reachable and upgraded by your proxy
- Configure TLS/CORS at the proxy or host level

---

## FAQ

**How do I opt-in per component?** Use a host signal type (`t.HInt`, `t.HString`, etc.) on a struct field. The type auto-calls `AddHostComponent`.

**Can I mix SSC and client-only components?** Yes, only components with host-type fields participate; others remain client-only.

**Where do `h:` values come from?** From the host component's handler; pushed over WebSocket and exposed in templates as `{h:key}`.

**Can the host trigger a client action?** Yes, by updating `h:` values or responding to `h:` commands.

---

## See Also

- [Manifest](../manifest), `rfw.json` reference
- [Host API](../api/host) & [HostClient API](../api/hostclient)
- [SSC API](../api/ssc)