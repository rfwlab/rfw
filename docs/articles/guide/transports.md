# SSC transports

RFW keeps the SSC API independent from its network transport. Components,
`hostclient.Send`, typed actions, hydration, session resumption and broadcasts
behave the same with WebSocket and Warp StreamBus.

## Selecting StreamBus

Set the top-level `transport` key in `rfw.json`:

```json
{
  "transport": "streambus",
  "build": {
    "type": "ssc",
    "host": ""
  }
}
```

The environment overrides the manifest:

```bash
RFW_TRANSPORT=streambus rfw dev
```

Accepted values are:

| Value | Behaviour |
| --- | --- |
| `websocket` | Use the existing `/ws` endpoint. This remains the default. |
| `streambus` | Prefer Warp StreamBus over WebTransport and fall back to WebSocket if the browser or network cannot establish it. |
| `auto` | Probe StreamBus first and transparently use WebSocket when unavailable. |

`webtransport` and `warp-streambus` are accepted aliases for `streambus`.

## Runtime model

StreamBus runs on WebTransport over HTTP/3 at `/streambus`. RFW preserves its
existing JSON SSC protocol and length-prefixes messages on a reliable QUIC
stream. On the server, Warp StreamBus provides bounded queues, priorities,
replay storage and explicit backpressure before frames reach the network.

The WebSocket endpoint remains mounted as a compatibility fallback. No
application code or component API changes when the selected transport changes.

Production deployments must route TCP and UDP for the HTTPS port and use a
certificate trusted by the browser. During `rfw dev`, the HTTP/3 endpoint uses
the HTTPS development port, which is the configured HTTP port plus one. RFW
automatically exposes the short-lived development certificate hash to the WASM
client so WebTransport can validate the self-signed certificate without an
application-level exception.

Set `window.RFW_STREAMBUS_URL` before loading the WASM module only when the
HTTP/3 endpoint uses a different public URL. Otherwise RFW derives it from
`build.host` or the page origin.

## Performance scope

The checked-in fan-out benchmark measures producer-side dispatch cost. It
compares StreamBus publication into bounded subscriber queues with RFW's
synchronous WebSocket fallback send path. Production WebSocket endpoints also
have an outbound queue, so this is deliberately not an end-to-end latency or
network-throughput claim.

Run it with:

```bash
go test ./host -run '^$' -bench 'Benchmark(StreamBus|WebSocket)Fanout' -benchmem -benchtime=2s -count=3
```

These numbers do not claim that every individual WebTransport round trip is
faster than every WebSocket round trip. They demonstrate the improvement for
the heavy fan-out and burst workloads StreamBus is designed to handle.

Reference run on Go 1.25.0 and an AMD EPYC 9V74, median of three 500 ms runs:

| Subscribers | StreamBus | WebSocket | Producer speedup |
| ---: | ---: | ---: | ---: |
| 1 | 878 ns/op | 4,483 ns/op | 5.1x |
| 10 | 10,505 ns/op | 47,526 ns/op | 4.5x |
| 100 | 115,784 ns/op | 507,769 ns/op | 4.4x |

At 100 subscribers, StreamBus used 28,801 B and 500 allocations per broadcast,
compared with 112,964 B and 1,801 allocations for synchronous WebSocket
fan-out. That is about 3.9x less allocated memory and 3.6x fewer allocations on
the producer path. The StreamBus benchmark includes the distinct SSC JSON
serialization and queue publish required for every target session.
