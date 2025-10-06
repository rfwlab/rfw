# host

The `host` package lets Go code back HTML components with server side logic. Components register handlers and communicate with the browser over a WebSocket.

| Item | Description |
| --- | --- |
| `HostComponent` | Couples a component name with a handler. |
| `Register(hc *HostComponent)` | Adds a component to the global registry. |
| `HandlerWithSession` | Function signature `func(*Session, map[string]any) any`. |
| `NewHostComponentWithSession(name string, handler HandlerWithSession) *HostComponent` | Registers a session-aware handler that receives per-connection context. |
| `(*HostComponent).HandleWithSession(session *Session, payload map[string]any) any` | Executes the session-aware callback, falling back to legacy handlers. |
| `ListenAndServe(addr, root string)` | Serves files and exposes the WebSocket endpoint. |
| `NewMux(root string)` | Returns a configured `*http.ServeMux`. |
| `Broadcast(name string, payload any, opts ...BroadcastOption)` | Sends a payload to all subscribers or to a filtered session set. |
| `WithSessionTarget(sessionID string) BroadcastOption` | Restricts a broadcast to a single session ID. |
| `SessionByID(id string) (*Session, bool)` | Looks up the session currently attached to a WebSocket. |
| `(*Session).ID() string` | Returns the server-generated session identifier. |
| `(*Session).StoreManager() *state.StoreManager` | Exposes an isolated store manager for the connection. |
| `(*Session).ContextSet/Get/Delete` | Manage arbitrary per-session context data. |
| `(*Session).Snapshot() map[string]map[string]map[string]any` | Captures all stores registered in the session. |

See the [SSC guide](../guide/ssc.md#session-scoped-hydration-data) for usage patterns.

