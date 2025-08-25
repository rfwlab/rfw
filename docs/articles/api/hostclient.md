# hostclient

The `hostclient` runtime runs inside the WebAssembly bundle and maintains a WebSocket connection to the host process.

## RegisterComponent

`RegisterComponent(id, name string, vars []string)` binds an HTML component instance to a server side `HostComponent`. It stores the element id and list of host variables, opens the socket if necessary, and sends an initial message so the host can respond with variable values.

## Send

`Send(name string, payload any)` serialises the payload and writes it to the socket. The message targets the host component by name.

## Variable bindings

When messages come back from the host, `hostclient` finds elements marked with `data-host-var` and updates their `textContent`. This keeps variables prefixed with `h:` in sync with server state.
