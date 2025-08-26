package host

import (
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/websocket"
)

// NewMux returns an HTTP mux that serves static files from root and the
// WebSocket handler at /ws.
func NewMux(root string) *http.ServeMux {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(root))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(root, r.URL.Path)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
	mux.Handle("/ws", websocket.Handler(wsHandler))
	return mux
}

// ListenAndServe starts an HTTP server using NewMux to serve files and the
// WebSocket endpoint.
func ListenAndServe(addr, root string) error {
	return http.ListenAndServe(addr, NewMux(root))
}
