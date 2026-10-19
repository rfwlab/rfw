# SSC transports

Components, `hostclient.Send`, typed actions, hydration, resumption and
broadcasts share the SSC message format over WebSocket and Warp StreamBus.
Browser support and credential requirements differ between the transports.

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
and explicit backpressure before frames reach the network. SSC keeps its own
sequenced replay history above the transport.

Browser StreamBus sends enqueue whole frames and return without waiting for a
JavaScript stream promise, including when called inside a DOM event handler.
One writer preserves frame order. Each connection admits at most 256 pending
frames and 8 MiB of JSON payload, including its in-flight frame. Exceeding either
limit or failing a native write ends the connection so SSC can reconnect and
replay its unacknowledged outbox. Enqueue success is not a delivery acknowledgement.

Browser constructor, readiness and stream-setup failures return to the same
fallback path. A partially created WebTransport is closed before fallback;
a synchronous JavaScript exception must not escape the connection attempt.
Application retry and authorization policies remain the caller's responsibility.

## Cookie authentication

Browser WebTransport uses
[credentials mode `omit`](https://www.w3.org/TR/webtransport/#dom-webtransport-webtransport)
and does not send the application's session cookie. A cookie-based
`WithAuthFunc` therefore rejects its upgrade; compatibility mode can fall back
to an authenticated WebSocket. A working greeting on that fallback does not
qualify authenticated StreamBus.

Opt in to the first-frame authentication exchange on the host:

```go
mux := host.NewMux(root,
    host.WithAuthFunc(validCookie),
    host.WithSSCStreamBusAuth(),
    host.WithSSCSessionInitializer(initializeIdentity),
    host.WithSSCResumeAuthorizer(authorizeResume),
)
```

The browser POSTs an empty `application/json` request to the same-origin
`/__rfw/streambus-auth` endpoint with its normal HTTP credentials. The response
is not cacheable. It holds a random single-use credential, not the cookie.
The WASM client sends that credential in the first stream frame and waits for
authentication acknowledgement before sending SSC messages. No credential is
placed in the URL or browser storage.

Credentials expire after 30 seconds and bind to the browser's exact Origin.
Each endpoint retains at most 256 outstanding credentials with at most 16 KiB
of request headers per credential. Consumption is atomic and rechecks
`WithAuthFunc` against the captured identity and actual upgrade URL. The
initializer and resume authorizer receive the authenticated headers. Identity
middleware must work on both the bootstrap and upgrade paths; authorization
must consult current server state rather than trusting cached identity claims.

The option admits an origin-checked transport before identity is known, then
requires authentication before creating an SSC session. Accepting a stream,
reading its at-most-4-KiB authentication frame and writing acknowledgement have
five-second bounds under the connection limit. Missing, expired, reused or
revoked credentials close the transport. Existing consumers retain the original
upgrade guard unless this option is enabled. Configure an explicit origin
allowlist for reverse proxies or a separately hosted client; the bootstrap also
requires its exact request origin or an explicitly configured allowed origin.

Deploy the host and WASM client together before enabling the option. Older
clients do not send the authentication frame and cannot use that opted-in
endpoint; reload old tabs after the upgrade. The default remains unchanged.

This exchange authenticates connection establishment. It refuses resumption
unless `WithSSCResumeAuthorizer` is configured, including on the WebSocket
fallback of that mux. Keep message/action authorization configured. When changing
identities, unmount user-owned scopes, update the HTTP authentication state,
call `hostclient.ResetSession()`, then mount the replacement scope. Reset ends
pending typed calls with `ErrSessionReset` and reconnects without waiting for
the old transport's remote close acknowledgement. Apply asynchronous results
only to their still-mounted owner. Reset does not cancel work already executing
on the host or turn the in-memory SSC session into durable application storage.

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

## Stock host startup

`host.Start` binds the HTTP port, HTTPS port and, when enabled, the UDP
StreamBus port before serving requests. A bind failure returns an error rather
than leaving only one transport running. A later serve failure stops the sibling
HTTP servers and HTTP/3 listener. This is error cleanup, not an application
shutdown or in-flight request draining API.

The generated host uses a short-lived P-256 self-signed certificate. Certificate
hash validation [requires P-256 support and excludes RSA keys](https://www.w3.org/TR/webtransport/#certificate-hashes).
HTTP/3 advertises `h3` through its TLS configuration. The same mux origin and
identity guards protect WebSocket and StreamBus, including when the HTTP frontend
connects to the adjacent HTTPS port. Configure an origin allowlist and identity
guard for authenticated deployments; the default mux permits any origin.

A successful SSC greeting alone does not prove StreamBus was selected: it may
have arrived over the documented WebSocket fallback. Browser acceptance must
observe a ready WebTransport connection and SSC traffic without WebSocket frames.
