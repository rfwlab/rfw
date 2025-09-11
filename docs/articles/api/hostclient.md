# hostclient

The `hostclient` runtime runs inside the WebAssembly bundle and maintains a WebSocket connection to the host process.
It automatically selects `ws` or `wss` based on the page's protocol.

## RegisterComponent

`RegisterComponent(id, name string, vars []string)` binds an HTML component instance to a server side `HostComponent`. It stores the element id and list of host variables, opens the socket if necessary, and sends an initial message so the host can respond with variable values.

## Send

`Send(name string, payload any)` serialises the payload and writes it to the socket. The message targets the host component by name.

## RegisterHandler

`RegisterHandler(name string, h func(map[string]any))` attaches a callback for inbound payloads addressed to `name`. It opens the WebSocket if needed and sends an initial message so the host can start broadcasting to this client. Use this when messages should be consumed by Go code instead of updating DOM bindings.

## EnableDebug

`EnableDebug()` writes WebSocket connection and message events to the browser console. Invoke it before sending or receiving messages when you need to inspect traffic.

## Variable bindings

When messages come back from the host, `hostclient` finds elements marked with `data-host-var` and updates their `textContent`. This keeps variables prefixed with `h:` in sync with server state.
