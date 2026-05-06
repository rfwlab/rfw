# netcode

```go
import "github.com/rfwlab/rfw/v2/netcode"
```

Networking utilities for multiplayer games. Client-side prediction and server reconciliation.

## Client

```go
type Client[T any] struct { ... }
```

Network client with command queue and interpolation.

| Method | Description |
| --- | --- |
| `Enqueue(cmd any)` | Queue a command for next flush |
| `Flush(tick int64)` | Send queued commands with tick |
| `State() T` | Get interpolated state |
| `Peek() T` | Get latest server state (no interpolation) |

### Constructor

```go
func NewClient[T any](name string, decode func(map[string]any) T, interp func(T, T, float64) T) *Client[T]
```

- `decode`: converts server state payload to T
- `interp`: interpolates between two states at factor t (0-1)

## Server

```go
type Server struct {
    Addr       string
    Games      map[string]*Game
    TickRate   int
}
```

| Method | Description |
| --- | --- |
| `NewServer(addr string) *Server` | Create server |
| `Start()` error | Start listening |
| `Broadcast(game string, tick int64, state any)` | Broadcast state to all clients |

## Game

```go
type Game struct {
    Clients map[string]*ClientHandle
}
```

| Method | Description |
| --- | --- |
| `AddClient(id string)` | Add client to game |
| `RemoveClient(id string)` | Remove client |
| `Enqueue(id string, cmd any)` | Queue command for client |

## Example

```go
type PlayerState struct {
    X, Y float32
    Rot  float32
}

decode := func(m map[string]any) PlayerState {
    return PlayerState{
        X:  float32(m["x"].(float64)),
        Y:  float32(m["y"].(float64)),
        Rot: float32(m["rot"].(float64)),
    }
}

interp := func(a, b PlayerState, t float64) PlayerState {
    return PlayerState{
        X:  a.X + (b.X - a.X)*t,
        Y:  a.Y + (b.Y - a.Y)*t,
        Rot: a.Rot + (b.Rot - a.Rot)*t,
    }
}

client := netcode.NewClient[PlayerState]("player", decode, interp)
```