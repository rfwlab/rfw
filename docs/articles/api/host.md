# host

The `host` package lets Go code back HTML components with server-side logic. Components register handlers and communicate with the browser over WebSocket.

---

## Struct-Based Host Components

Define a struct implementing the `host.Component` interface:

```go
type Component interface {
    Name() string
    Serve(host.Handler)
}
```

Then register it with `host.RegisterComponent`:

```go
type VisitComponent struct{}

func (v *VisitComponent) Name() string { return "Visit" }

func (v *VisitComponent) Serve(h host.Handler) {
    h.On("UpdateVisits", func(host.Payload) host.Payload {
        visits++
        return host.Payload{"visit": visits}
    })
}

func main() {
    host.RegisterComponent(&VisitComponent{})
    // ...
}
```

### Handler Methods

The `host.Handler` passed to `Serve` provides:

| Method | Description |
| --- | --- |
| `On(cmd string, fn func(Payload) Payload)` | Register a command handler |
| `OnSession(cmd string, fn func(*Session, Payload) Payload)` | Register a session-aware command handler |
| `WithInitSnapshot(fn func(*Session, map[string]any) *InitSnapshot)` | Set initial snapshot callback |

---

## Function-Based Host Components

For simple cases, use `NewHostComponent` or `NewHostComponentWithSession`:

| Function | Description |
| --- | --- |
| `NewHostComponent(name string, handler func(map[string]any) any)` | Creates a host component with a simple handler |
| `NewHostComponentWithSession(name string, handler func(*Session, map[string]any) any)` | Creates a session-aware host component |

---

## HTML Helpers

The `host` package provides HTML builders for server-side rendering with automatic `data-host-var` and `data-host-expected` attributes:

```go
import "github.com/rfwlab/rfw/v2/host"

h.Span("Visit", host.Var("visit"), host.Expected("visit"))
h.Div(h.P("Content"))
h.Tag("section", h.Text("Hello"))
```

| Helper | Description |
| --- | --- |
| `Span(children...)` | `<span>` element |
| `Div(children...)` | `<div>` element |
| `P(children...)` | `<p>` element |
| `Tag(name, children...)` | Generic element |
| `Text(s)` | Text node |
| `Raw(html)` | Raw HTML |
| `Join(nodes...)` | Join multiple nodes |
| `Var(name)` | Mark a `data-host-var` binding |
| `Expected(name)` | Mark a `data-host-expected` binding |

---

## Starting the Server

| Function | Description |
| --- | --- |
| `ListenAndServe(addr, root string)` | Serves files and exposes the WebSocket endpoint |
| `NewMux(root string)` | Returns a configured `*http.ServeMux` |
| `StartAuto()` | Reads `rfw.json` config, falls back to `"build/client"` |

---

## Registration

| Function | Description |
| --- | --- |
| `Register(hc *HostComponent)` | Adds a component to the global registry |
| `Get(name string) (*HostComponent, bool)` | Looks up a registered component |
| `RegisterComponent(c Component)` | Registers a struct-based host component |

---

## Broadcasting

| Function | Description |
| --- | --- |
| `Broadcast(name string, payload any, opts ...BroadcastOption)` | Sends a payload to all subscribers or filtered sessions |
| `WithSessionTarget(sessionID string) BroadcastOption` | Restricts broadcast to a single session ID |

---

## Session

| Method | Description |
| --- | --- |
| `(*Session).ID() string` | Returns the server-generated session identifier |
| `(*Session).StoreManager() *state.StoreManager` | Exposes an isolated store manager for the connection |
| `(*Session).ContextSet/Get/Delete` | Manage arbitrary per-session context data |
| `(*Session).Snapshot() map[string]map[string]map[string]any` | Captures all stores registered in the session |

---

## Payload

`host.Payload` is `map[string]any` — the data exchanged between client and server over WebSocket.

See the [SSC guide](../guide/ssc.md) for usage patterns.