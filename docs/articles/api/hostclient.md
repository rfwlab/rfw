# hostclient

The `hostclient` runtime runs inside the WebAssembly bundle and maintains a WebSocket connection to the host process. It automatically selects `ws` or `wss` based on the page's protocol.

| Function | Description |
| --- | --- |
| `RegisterComponent(id, name string, vars []string)` | Bind an element to a server-side `HostComponent`. |
| `Send(name string, payload any)` | Serialise and write a payload to the socket. |
| `RegisterHandler(name string, h func(map[string]any))` | Attach a callback for inbound payloads. |
| `EnableDebug()` | Log WebSocket connection and message events. |

