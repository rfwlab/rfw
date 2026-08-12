package host

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"expvar"
	"math/big"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

// ResolveRoot resolves a content root relative to the executable when needed.
func ResolveRoot(root string) string {
	if _, err := os.Stat(root); err == nil {
		return root
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", root)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return root
}

// NewMux returns an HTTP mux that serves static files from root and the
// WebSocket handler at /ws. Options gate the WebSocket endpoint; by default it
// accepts any origin and identity.
func NewMux(root string, opts ...MuxOption) *http.ServeMux {
	root = ResolveRoot(root)
	runtime := NewWSRuntime(opts...)
	staticRoot := filepath.Join(root, "..", "static")
	mux := http.NewServeMux()
	if os.Getenv("RFW_DEVTOOLS") != "" {
		mux.Handle("/debug/vars", expvar.Handler())
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	fs := http.FileServer(http.Dir(root))
	rootDir := http.Dir(root)
	var sfs http.Handler
	var staticDir http.Dir
	if _, err := os.Stat(staticRoot); err == nil {
		staticDir = http.Dir(staticRoot)
		sfs = http.FileServer(staticDir)
	}
	if sfs != nil {
		mux.Handle("/static/", http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setWasmEncodingHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
			sfs.ServeHTTP(w, r)
		})))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if sfs != nil {
			if regularFile(staticDir, r.URL.Path) {
				setWasmEncodingHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
				sfs.ServeHTTP(w, r)
				return
			}
		}
		if regularFile(rootDir, r.URL.Path) {
			setWasmEncodingHeaders(w, r.URL.Path, r.URL.Query().Get("v") != "")
			fs.ServeHTTP(w, r)
			return
		}
		// Serve index.html only for HTML requests or bare paths to avoid
		// returning HTML for CSS, JS, image, etc. requests.
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") || r.URL.Path == "/" || r.URL.Path == "" {
			if devMode() {
				w.Header().Set("Cache-Control", "no-store")
			}
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		http.NotFound(w, r)
	})
	mux.Handle("/ws", runtime.Guard(websocket.Handler(func(ws *websocket.Conn) {
		wsHandler(ws, runtime)
	})))
	return mux
}

// ListenAndServe starts an HTTP server using NewMux to serve files and the
// WebSocket endpoint.
func ListenAndServe(addr, root string) error {
	logger.Info("serving HTTP", "addr", addr)
	return newHTTPServer(addr, loggingMiddleware(NewMux(root))).ListenAndServe()
}

// ListenAndServeWithMux starts an HTTP server using the provided mux.
func ListenAndServeWithMux(addr string, mux *http.ServeMux) error {
	logger.Info("serving HTTP", "addr", addr)
	return newHTTPServer(addr, loggingMiddleware(mux)).ListenAndServe()
}

// ListenAndServeTLS starts an HTTPS server using a self-signed certificate
// and NewMux to serve files and the WebSocket endpoint.
func ListenAndServeTLS(addr, root string) error {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return err
	}
	srv := newHTTPServer(addr, loggingMiddleware(NewMux(root)))
	srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	logger.Info("serving HTTPS", "addr", addr)
	return srv.ListenAndServeTLS("", "")
}

// ListenAndServeTLSWithMux starts an HTTPS server using a self-signed certificate
// and the provided mux, preserving any additional routes registered by callers.
func ListenAndServeTLSWithMux(addr string, mux *http.ServeMux) error {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return err
	}
	srv := newHTTPServer(addr, loggingMiddleware(mux))
	srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	logger.Info("serving HTTPS", "addr", addr)
	return srv.ListenAndServeTLS("", "")
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func regularFile(root http.Dir, name string) bool {
	f, err := root.Open(name)
	if err != nil {
		return false
	}
	info, statErr := f.Stat()
	closeErr := f.Close()
	return statErr == nil && closeErr == nil && !info.IsDir()
}

// devMode reports whether the server is running under `rfw dev`. The dev command
// exports RFW_DEV_BUILD=1 and propagates it to the SSC host child via os.Environ,
// so both the static and host-proxied serving paths observe it.
func devMode() bool { return os.Getenv("RFW_DEV_BUILD") == "1" }

func setWasmEncodingHeaders(w http.ResponseWriter, path string, versioned bool) {
	// In dev, nothing may be cached: the wasm version pointer lives in
	// rfw_config.js and the binary is fetched as app.wasm?v=<hash>. Caching
	// either one leaves the browser re-requesting a stale ?v= against an
	// immutable entry, so rebuilds are never picked up. no-store on every asset
	// forces a fresh fetch each load. Production keeps the immutable policy.
	if devMode() {
		w.Header().Set("Cache-Control", "no-store")
	}
	if !strings.HasSuffix(path, ".wasm") && !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	header := w.Header()
	if !devMode() {
		if versioned {
			header.Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			header.Set("Cache-Control", "no-cache")
		}
	}
	if !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	header.Set("Content-Encoding", "br")
	header.Set("Content-Type", "application/wasm")
	if vary := header.Get("Vary"); vary == "" {
		header.Set("Vary", "Accept-Encoding")
	} else if !strings.Contains(vary, "Accept-Encoding") {
		header.Set("Vary", vary+", Accept-Encoding")
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return tls.X509KeyPair(certPEM, keyPEM)
}
