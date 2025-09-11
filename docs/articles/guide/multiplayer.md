# Multiplayer

## Why

Real-time games often need to keep client state in sync with a server while hiding network jitter. The `netcode` package layers snapshot interpolation and command queues on top of the existing WebSocket channel.

## When to use

Use `netcode` when multiple clients interact with shared state that must stay authoritative on the server. It is not intended for large persistent worlds.

## Setup

1. Register a `netcode.Server` as a host component and start broadcasting snapshots.
2. Create a `netcode.Client` inside your Wasm component to enqueue commands and interpolate snapshots. The handler registration opens the WebSocket and subscribes to server snapshots automatically.
3. Call `Flush` on a fixed interval to send queued commands before reading the interpolated state.

## Message formats

### Commands → server

```json
{
  "tick": 100,
  "commands": [ { "dx": 5 } ]
}
```

### Snapshots → clients

```json
{
  "tick": 100,
  "state": { "x": 5 }
}
```

## API used

- `netcode.NewServer(name, initial, apply)`
- `(*Server).Broadcast(tick)`
- `netcode.NewClient[T](name, decode, interp)`
- `(*Client).Enqueue(cmd)`
- `(*Client).Flush(tick)`
- `(*Client).State(now)`
- `host.Broadcast(name, payload)`
- `hostclient.Send(name, payload)`
- `hostclient.RegisterHandler(name, handler)`

## Server example

```go
srv := netcode.NewServer("Game", testState{}, func(s *testState, cmd any) {
    m := cmd.(map[string]any)
    s.X += m["dx"].(float64)
})
host.Register(srv.HostComponent())
```

## Client example

```go
c := netcode.NewClient[testState]("Game", decodeState, lerp)
go func() {
    ticker := time.NewTicker(50 * time.Millisecond)
    var tick int64
    for range ticker.C {
        tick += 50
        c.Flush(tick)
        state := c.State(tick)
        fmt.Println(state)
    }
}()
c.Enqueue(map[string]any{"dx": 5})
```

Explore the demo below.

@include:ExampleFrame:{code:"/examples/components/netcode_component.go", uri:"/examples/netcode"}

## Debugging

Call `hostclient.EnableDebug()` before constructing the client to log WebSocket events to the browser console:

```go
hostclient.EnableDebug()
c := netcode.NewClient[testState]("Game", decodeState, lerp)
```

## Limitations

- No built-in entity reconciliation.
- All payloads must be JSON serialisable.

## Related links

- [host](../api/host)
- [hostclient](../api/hostclient)
