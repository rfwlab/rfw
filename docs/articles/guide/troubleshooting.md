# Troubleshooting from observed boundaries

Start with the failing command or browser request, its status and the selected
CLI/framework revision. Keep the actual response body and a minimal reproduction.
Do not classify a browser failure by a vendor-specific exception string or assume
that a successful compile proves the application mounted.

| Symptom | Check | Next action |
| --- | --- | --- |
| `rfw` is not found | The selected Go binary's GOBIN/GOPATH and shell PATH | Run the installed executable by its full path; avoid accidentally testing another CLI version. |
| A fresh scaffold downloads an unexpected SDK | The CLI version and generated go.mod | Candidate CLIs pin their own framework version. A local replacement tests that checkout, not a released package. Review the generated module before building. |
| Readonly build wants module updates | go.mod/go.sum, replaced module and selected Go version | Run tidy in the controlled source workspace, review and commit the lock changes, then repeat the readonly build. Do not silently remove the readonly check. |
| A page exists but does not route | rfw.json plugins and generated pages registration | Use `rfw build` or `rfw dev` to run the pages plugin. An explicit empty plugins map disables defaults. Plain Go compilation does not generate routes. |
| WASM cannot instantiate | Asset URL, status, MIME, content coding and decoded bytes | Compare hashes with the complete built generation. HTML from an SPA fallback and unlabelled brotli are not WASM. Follow the loader's visible failure rather than retrying arbitrary files. |
| Reload keeps an older bundle | index/config cache headers and versioned bootstrap URLs | Revalidate entry/config, deploy one generation and check proxy caches. Do not mark the version pointer immutable. Keep prior assets or enforce a maintenance reload policy. |
| Assets are repeatedly fetched | RFW_DEV_BUILD and version query | Development deliberately uses no-store. Unset the environment flag for a production acceptance fixture. |
| A cached HTTP error never recovers | FetchJSON/FetchText cache key | These helpers retain errors. Explicitly ClearCache for that URL before a reviewed retry; use Request for uncached mutations and authenticated requests. |
| A form stays busy after failure | Whether every HTTP status/body branch completes local state | Request completes transport/browser/body-read failures with status 0. It never retries a write. Keep the draft and reconcile server state before resubmitting. |
| A late response overwrites another page | Which mount context the callback captured | Check the captured context before touching state/DOM. A fresh Scope inside the callback can belong to a later mount. Request currently has no abort handle. |
| StreamBus config returns 503 | Selected transport and whether HTTP/3 certificate/listener setup completed | Registering the mux is insufficient. Use a server path that prepares the HTTP/3 endpoint and inspect its bind/start error. |
| StreamBus returns 426 on ordinary HTTP | Whether the request is an actual WebTransport upgrade over HTTP/3 | Plain GET/CONNECT on an HTTP/1 fixture is not a transport success. Validate the real secure UDP endpoint. |
| Upgrade returns401/403 | Actual Origin, identity callback and deployment allowlist | Fix the intended identity/origin configuration. Do not disable guards to make the handshake green. |
| SSC greeting works but no WebTransport is observed | Actual connection type and fallback policy | WebSocket compatibility can carry the greeting. Check ready WebTransport and protocol traffic with WebSocket unavailable before claiming StreamBus acceptance. |
| Debug routes are exposed | RFW_DEVTOOLS value | Unset the variable. Every nonempty value, including 0, enables debug routes. |

## Evidence to keep

Record OS/architecture, Go version, module graph, CLI version, application revision,
manifest, build log and artifact hashes. In the browser, preserve the failing URL,
HTTP status/content headers, page error and reproduction steps. Remove credentials
and private payloads before sharing a report. If the failure requires a private
provider or production proxy, identify that untested boundary explicitly.

Reproduce first with the nearest runnable example. Counter isolates mounting and
events; dynamic-list isolates delegated runtime markup; dashboard isolates feed
lifecycle; Notes isolates real HTTP CRUD/conflicts. SSC and StreamBus need their
own host, authority and network fixtures. A static example cannot certify those
paths merely because it renders in the same browser.
