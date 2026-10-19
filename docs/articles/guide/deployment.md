# Deploy the built application

RFW produces application artifacts. It does not configure your DNS, certificates,
process supervisor, database recovery or application authorization. Keep those
contracts explicit before exposing an application to users.

## Build one complete generation

Use the patched Go baseline in go.mod and a pinned CLI/framework revision. Build
from a committed application tree and retain its go.mod/go.sum and rfw.json.
`rfw build` produces build/client, build/static and, for an SSC project, its host
binary. See [bundle delivery](wasm-delivery.md) for compressed files and stamped
bootstrap tags. Do not deploy one file copied out of a still-running build.

Record the application revision, framework module version/checksum, Go version
and SHA256 hashes of the native executable and complete served asset tree. A
checksum identifies bytes; it does not prove their source, permissions or safety.
Retain previous generations and their matching configuration for rollback.

## Choose the serving shape

A static application can use a static host/CDN. Preserve generated file names,
MIME types and compression behavior. Serve index.html and rfw_config.js with
revalidation, not a permanent cache. The loader uses compressed named artifacts
for static delivery; the host must either label their encoding correctly or let
the loader decode the supported gzip artifact. Missing JS/CSS must return404,
not the SPA index. Configure route fallback only for HTML navigation.

A native application can use host.NewMux on a directory or host.NewMuxFS on an
fs.FS subtree. For embedding, include only the built public tree, then select its
subdirectory with fs.Sub. Do not serve a repository root that contains secrets,
source configuration or unrelated files. NewMux has a sibling static directory;
NewMuxFS's input is the complete served tree, so include any needed static assets
inside that tree. Both negotiate compressed WASM, revalidate SPA entries and
register the selected SSC transport. NewMuxFS does not run an HTTP/3 listener by
itself; the server startup path must supply one for StreamBus.

The generic mux permits SSC connections without application identity checks by
default. Configure WithAuthFunc and WithOriginAllowlist for your deployment.
Those options gate transport upgrades; they do not authorize every component,
action, entity or API route for you. Verify user/workspace permission again at
execution time. Do not treat a successful greeting as a security acceptance test.

## HTTP, HTTPS and StreamBus

StartAuto reads RFW_HOST_PORT or the manifest port, serves HTTP there and HTTPS
on the next port, and resolves its client root relative to the running executable
when appropriate. Its generated self-signed certificate is for development;
it is not trusted production TLS or a certificate renewal solution. It binds
both TCP listeners before serving and, when configured, prepares HTTP/3 on the
secure port. A bind error returns to the caller. The default Start path binds
all interfaces, so restrict access outside the process until application guards
and public-origin configuration are ready.

For an HTTP reverse proxy, construct the guarded mux and serve it on an explicitly
bound private address. Use your trusted TLS termination and proxy WebSocket
upgrades only when that compatibility path is offered. StreamBus requires a
reachable HTTP/3/WebTransport endpoint and certificate policy, not just an HTTP/1
proxy rule. See [transport selection](transports.md). Test UDP reachability,
origins, handshake failures and the actual chosen transport in supported browsers.
A fallback WebSocket greeting does not prove StreamBus worked.

Unset RFW_DEV_BUILD and RFW_DEVTOOLS in production. The former disables caching;
any nonempty RFW_DEVTOOLS value exposes debug/pprof routes, including the string
"0". Development watchers and rebuild servers are not process supervisors.
Use separate credentials and origins for acceptance fixtures and real users.

## Upgrade and rollback

Stop new traffic and the old process before replacing a native executable or
changing its asset root. Stage a complete generation beside the previous one,
verify its recorded hashes and switch the selected path atomically where your
platform supports that operation. Keep executable, loader, runtime config, raw
WASM and compressed variants from the same generation. An embedded executable
keeps that tree together; it does not migrate your backend data.

A content-versioned WASM URL is immutable for cache purposes. Keep old generations
available for clients still loading them, or define a maintenance/reload policy
that prevents mixed generations. RFW does not retain historical assets for you.
Verify a normal cached browser reload receives the new index/config version and
loads the matching bundle. Rollback requires the old application configuration
and compatible data, not just an older WASM file.

Back up application data and secrets with the application's own procedure before
schema-changing upgrades. RFW has no generic ORM, datastore migration or backend
rollback. Check the application's release notes and test restoration on an
isolated copy. Do not infer schema compatibility from a framework version number.

## Acceptance checklist

Check actual listener startup and failure on an occupied port; correct root and
asset hashes; gzip/brotli/raw responses; GET and HEAD headers; SPA revalidation;
missing asset404; transport origin/identity refusal; the real UI in each supported
browser; cached reload after a generation change; and shutdown/restart under the
chosen supervisor. Record which checks ran on a local fixture and which still
need the actual public proxy, certificates, UDP network and deployment platform.
