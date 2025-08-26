# host

The `host` package lets Go code back HTML components with server side logic. Components register handlers and communicate with the browser over a WebSocket.

## HostComponent

A `HostComponent` couples a component name with a `Handler`:

```go
 type Handler func(payload map[string]any) any
 type HostComponent struct {
     name    string
     handler Handler
 }
```

The handler receives payloads sent from the client and may return a response payload to push back through the socket.

## Register

`Register(hc *HostComponent)` adds a component to the global registry so the WebSocket can route incoming messages to it.

## WebSocket server

`ListenAndServe(addr, root string)` starts an HTTP server that serves files from `root` and exposes the WebSocket endpoint at `/ws`. Messages arriving on the socket are dispatched to the matching `HostComponent` and the response, if any, is sent back to the caller. Use `Broadcast` to push a payload to all clients subscribed to a component.

`NewMux(root string)` returns an `*http.ServeMux` preconfigured in the same way, allowing additional handlers to be registered before calling `http.ListenAndServe`.
