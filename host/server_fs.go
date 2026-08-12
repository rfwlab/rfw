package host

import (
	"expvar"
	"io/fs"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

// NewMuxFS is the fs.FS counterpart of NewMux: it serves the client build from
// an fs.FS (for example an embed.FS sub-tree) instead of a directory on disk,
// and registers the WebSocket handler at /ws. This lets an application ship as
// a single self-contained binary with the build embedded via go:embed, or mount
// the rfw endpoints on assets it already holds in memory.
//
// fsys is treated as the complete served tree: index.html, app.wasm and any
// static assets must live inside it. Unlike NewMux there is no on-disk sibling
// static directory. Options gate the WebSocket endpoint exactly as in NewMux.
func NewMuxFS(fsys fs.FS, opts ...MuxOption) *http.ServeMux {
	runtime := NewWSRuntime(opts...)
	mux := http.NewServeMux()
	if os.Getenv("RFW_DEVTOOLS") != "" {
		mux.Handle("/debug/vars", expvar.Handler())
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	fileServer := http.FileServerFS(fsys)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if regularFileFS(fsys, r.URL.Path) {
			setWasmEncodingHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
			fileServer.ServeHTTP(w, r)
			return
		}
		// Serve index.html only for HTML requests or bare paths, so a missing
		// asset (CSS, JS, image) still returns 404 instead of HTML.
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") || r.URL.Path == "/" || r.URL.Path == "" {
			serveIndexFS(w, r, fsys)
			return
		}
		http.NotFound(w, r)
	})
	mux.Handle("/ws", runtime.Guard(websocket.Handler(func(ws *websocket.Conn) {
		wsHandler(ws, runtime)
	})))
	return mux
}

// serveIndexFS writes fsys's index.html, the SPA entry point.
func serveIndexFS(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	http.ServeFileFS(w, r, fsys, "index.html")
}

// regularFileFS reports whether name (a URL path) maps to a regular file in
// fsys, the fs.FS analogue of regularFile.
func regularFileFS(fsys fs.FS, name string) bool {
	name = strings.TrimPrefix(name, "/")
	if name == "" || !fs.ValidPath(name) {
		return false
	}
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}
