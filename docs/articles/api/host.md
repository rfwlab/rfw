# host

The `host` package lets Go code back HTML components with server side logic. Components register handlers and communicate with the browser over a WebSocket.

| Item | Description |
| --- | --- |
| `HostComponent` | Couples a component name with a handler. |
| `Register(hc *HostComponent)` | Adds a component to the global registry. |
| `ListenAndServe(addr, root string)` | Serves files and exposes the WebSocket endpoint. |
| `NewMux(root string)` | Returns a configured `*http.ServeMux`. |

