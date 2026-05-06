# ssc

```go
import "github.com/rfwlab/rfw/v2/ssc"
```

Server-Side Computed. WebSocket server enabling server-side rendering with reactive client updates.

## Quick Start

```go
server := ssc.NewSSCServer(":8080", "./dist")
log.Fatal(server.ListenAndServe())
```

## SSCServer

| Method | Description |
| --- | --- |
| `NewSSCServer(addr, root string) *SSCServer` | Create server with address and dist root. |
| `ListenAndServe() error` | Start HTTP server with WebSocket upgrade. |

## Events

```go
func SubscribeSSC(fn func(SSCEvent), priority ...events.Priority)
func EmitSSC(ctx context.Context, event SSCEvent) error
```

Subscribe to component events from clients.

## SSCEvent

```go
type SSCEvent struct {
    Component string
    Payload   map[string]any
    Session   *host.Session
}
```

## Broadcast

```go
func Broadcast(component string, payload any, opts ...host.BroadcastOption)
```

Send updates to all connected clients of a component. Options:
- `host.BroadcastToSession(id)` - send to specific session only

## API Routes

| Path | Description |
| --- | --- |
| `/` | Serves `index.html` or static files. |
| `/ws` | WebSocket endpoint for client connections. |
| `/static/` | Static files from `./static` directory. |

## WASM Headers

Server automatically sets `Content-Type` and correct headers for `.wasm` and `.wasm.br` files.