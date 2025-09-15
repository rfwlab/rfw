# netcode

Client and server helpers for real-time state synchronisation over WebSockets.

| Function | Description |
| --- | --- |
| `NewServer[T](name string, initial T, apply func(*T, any)) *Server[T]` | Create a server and register a host component. |
| `(s *Server[T]) HostComponent()` | Expose the host component for embedding. |
| `(s *Server[T]) Broadcast(tick int64)` | Send queued state to clients. |
| `NewClient[T](name string, decode func(map[string]any) T, interp func(T, T, float64) T) *Client[T]` | Create a client that decodes and interpolates snapshots. |
| `(c *Client[T]) Enqueue(cmd any)` | Queue a command to send to the server. |
| `(c *Client[T]) Flush(tick int64)` | Send queued commands. |
| `(c *Client[T]) State(now int64) T` | Get interpolated state at time `now`. |

