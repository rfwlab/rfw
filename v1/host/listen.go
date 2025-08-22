package host

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// ListenAndServe starts an HTTP server with the WebSocket handler registered at /ws.
func ListenAndServe(addr string) error {
	http.Handle("/ws", websocket.Handler(wsHandler))
	return http.ListenAndServe(addr, nil)
}
