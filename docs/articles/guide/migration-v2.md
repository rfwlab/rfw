# Updating a v2 consumer

This page covers the current candidate's verified compatibility changes. It is
not a published release announcement or a guarantee that every v2 application can
update without review. Select an approved source revision/release and keep the
previous application artifacts and configuration before starting.

## Record the actual dependency

A CLI and an application can use different framework versions. Record both:

```sh
rfw version
go list -m -json github.com/rfwlab/rfw/v2
go version
```

The CLI uses its installed framework version when scaffolding. Inspect the new
project's go.mod rather than assuming it chose the latest tag. Examples in this
repository use a local replacement: they prove checkout compatibility, not a
clean released-module installation. Remove a replacement only as a deliberate
consumer change, then tidy/review the graph and rebuild without that replacement.

Use the patched toolchain baseline in go.mod. Candidate/example modules currently
select x/net 0.60.0; retain the matching go.sum entries. Build native and js/wasm
surfaces actually used by the application. Some browser-only packages or methods
are not native server APIs, even though the module as a whole has native targets.

## Browser HTTP failures and cached values

Request and RequestBytes now complete browser/transport/body-read exceptions once
with status 0 and an empty body. The callback can run synchronously when a browser
call throws before returning a promise. Set local request ownership/busy state
before calling Request. Check status before decoding JSON and clear busy state
on every result. Callbacks are released on completion; application callback panics
still use the normal framework error path.

FetchJSON/FetchText reject HTTP error responses rather than treating their bodies
as successful application data. Their cache can retain that error; explicitly
ClearCache before a reviewed retry. Browser JSON is decoded in Go, preserving
integer fields instead of routing them through JavaScript number conversion.
Tests must assert actual error handling, not just an error-shaped object.

These requests do not automatically retry and do not expose an abort handle.
For writes, missing acknowledgement is not proof of rollback. Keep drafts and
reconcile through the application's idempotency/version contract. Capture the
original mount's context to ignore late responses after unmount.

## RTML event attributes and lifecycle

Only whole, valueless event attributes expand. Directive-shaped text in a URL,
attribute value, comment or raw-text body stays literal. Keep event declarations
as tag attributes and escape untrusted runtime text separately. Existing explicit
data-on attributes keep their syntax. Do not depend on the previous accidental
rewriting of arbitrary text.

Plain HTMLComponent values now run configured mount/unmount callbacks even when
SetComponent was not called. An application that manually invoked those callbacks
to compensate can now perform duplicate work. Remove that workaround and verify
that subscription/feed/widget initialization and cleanup occur once per mount.
Explicit component registration still uses the same lifecycle contract.

## Embedded deployment and transports

NewMuxFS now matches directory-backed WASM negotiation, entry revalidation and
selected StreamBus registration. Existing origin/identity options guard both
transports. Default WebSocket mode remains the default; selecting StreamBus can
now activate the embedded route that was previously absent. Provide the matching
HTTP/3 listener, TLS/certificate policy and UDP reachability before advertising it.
A mux alone is not a transport server.

Preserve the complete client generation, regenerate its loader/config and retest
normal cached reload and rollback. See [deployment](deployment.md) and
[troubleshooting](troubleshooting.md). The current local embedded acceptance uses
two synthetic bundle generations on Linux and three browsers. It does not certify
an application's schema migrations, private integrations, public proxy, installed
container, every mobile target or stable release approval.

## Consumer acceptance

At minimum, rebuild both required targets and run application tests for HTTP
failure/status handling, large numeric JSON values, event literals, lifecycle
cleanup, route generation, asset loading and current authority at transport/action
boundaries. Add a real browser flow with writes and a restart/reload where the
application persists data. Keep measured performance regressions visible: the
current 1,000-row directive benchmark has a recorded roughly 16% local slowdown.
Do not carry a framework test result over as proof that a consumer's business
flow, security or deployment has passed.
