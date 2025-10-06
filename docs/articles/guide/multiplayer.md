# Multiplayer

The **netcode** package helps real-time games synchronize client and server state while hiding network jitter. It layers snapshot interpolation and command queues on top of WebSocket channels.

## When to Use

Use `netcode` when multiple clients interact with shared state that must remain authoritative on the server. It’s intended for small to medium games, not large persistent worlds.

## Setup

1. Register a `netcode.Server` as a host component and broadcast snapshots.
2. Create a `netcode.Client` in your Wasm component to enqueue commands and interpolate snapshots. The client connects automatically and subscribes to updates.
3. Call `Flush` on a fixed interval to send commands and update state.

## Message Formats

### Commands → Server

```json
{
  "tick": 100,
  "commands": [ { "dx": 5 } ]
}
```

### Snapshots → Clients

```json
{
  "tick": 100,
  "state": { "x": 5 }
}
```

## API Reference

* `netcode.NewServer(name, initial, apply)`
* `(*Server).Broadcast(tick)`
* `(*Server).Update(func(*State))`
* `netcode.NewClient[T](name, decode, interp)`
* `(*Client).Enqueue(cmd)`
* `(*Client).Flush(tick)`
* `(*Client).State(now)`
* `host.Broadcast(name, payload)`
* `hostclient.Send(name, payload)`
* `hostclient.RegisterHandler(name, handler)`

## Server Example

```go
srv := netcode.NewServer("Game", testState{}, func(s *testState, cmd any) {
    m := cmd.(map[string]any)
    s.X += m["dx"].(float64)
})
host.Register(srv.HostComponent())
```

## Client Example

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
c.Enqueue(map[string]any{"dx": 5.0})
```

@include\:ExampleFrame:{code:"/examples/components/netcode\_component.go", uri:"/examples/netcode"}

The following game demonstrates the use of the netcode component:

@include\:ExampleFrame:{code:"/examples/components/multiplayer\_component.go", uri:"/examples/multiplayer"}

## Debugging

Enable debug logging to print WebSocket events to the browser console:

```go
hostclient.EnableDebug()
c := netcode.NewClient[testState]("Game", decodeState, lerp)
```

## Limitations

* No built-in entity reconciliation
* Payloads must be JSON serializable
* Command values are decoded as `float64` (enqueue numbers as floats, e.g. `1.0`)

## Related

* [host](../api/host)
* [hostclient](../api/hostclient)
