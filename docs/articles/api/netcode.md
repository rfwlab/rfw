# netcode

Client and server helpers for real-time state synchronisation over WebSockets.

## Why

Networked games need to keep an authoritative state while hiding latency from players.

## When to Use

Use `netcode` for small real-time games where a host must process commands and broadcast snapshots.

## How

1. Create a server with `netcode.NewServer` and register its host component.
2. Construct a client with `netcode.NewClient` to enqueue commands and interpolate snapshots.
3. Call `Flush` on a fixed interval before reading the interpolated `State`.

## API

```go
package netcode

type Server[T any] struct{}

func NewServer[T any](name string, initial T, apply func(*T, any)) *Server[T]
func (s *Server[T]) HostComponent() *host.HostComponent
func (s *Server[T]) Broadcast(tick int64)

type Client[T any] struct{}

func NewClient[T any](name string, decode func(map[string]any) T, interp func(T, T, float64) T) *Client[T]
func (c *Client[T]) Enqueue(cmd any)
func (c *Client[T]) Flush(tick int64)
func (c *Client[T]) State(now int64) T
```

## Example

```go
srv := netcode.NewServer("Game", state{}, apply)
host.Register(srv.HostComponent())

c := netcode.NewClient[state]("Game", decode, lerp)
c.Enqueue(map[string]any{"dx": 5.0})
c.Flush(tick)
_ = c.State(tick)
```

## Notes and Limitations

- All payloads must be JSON serialisable.
- No built-in entity reconciliation.

## Related Links

- [Multiplayer guide](../guide/multiplayer)
- [host](./host)
- [hostclient](./hostclient)
