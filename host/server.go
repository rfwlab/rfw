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

// MuxOption configures optional behaviour of the mux returned by NewMux,
// currently the guards applied to the /ws WebSocket endpoint.
type MuxOption func(*wsGuardConfig)

type wsGuardConfig struct {
	authFunc func(*http.Request) bool
	origins  []string
}

// WithAuthFunc registers a callback invoked before the WebSocket upgrade at
// /ws. It receives the upgrade request (cookies, headers, URL) and returning
// false rejects the connection with 401 before a session is allocated. The
// default remains open: without this option any client can connect.
func WithAuthFunc(fn func(*http.Request) bool) MuxOption {
	return func(cfg *wsGuardConfig) { cfg.authFunc = fn }
}

// WithOriginAllowlist restricts /ws upgrades to requests whose Origin header
// exactly matches one of the given origins (e.g. "https://app.example.com").
// Requests without an Origin header or with an unlisted one are rejected with
// 403. The default remains open: without this option any origin is accepted.
func WithOriginAllowlist(origins ...string) MuxOption {
	return func(cfg *wsGuardConfig) { cfg.origins = append(cfg.origins, origins...) }
}

// GuardWS wraps a WebSocket handler with the origin and auth checks configured
// through MuxOptions. It is used by NewMux and ssc.NewSSCServer to gate /ws.
func GuardWS(next http.Handler, opts ...MuxOption) http.Handler {
	var cfg wsGuardConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.authFunc == nil && len(cfg.origins) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.origins) > 0 {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range cfg.origins {
				if origin == o {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
		}
		if cfg.authFunc != nil && !cfg.authFunc(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewMux returns an HTTP mux that serves static files from root and the
// WebSocket handler at /ws. Options gate the WebSocket endpoint; by default it
// accepts any origin and identity.
func NewMux(root string, opts ...MuxOption) *http.ServeMux {
	root = ResolveRoot(root)
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
	var sfs http.Handler
	if _, err := os.Stat(staticRoot); err == nil {
		sfs = http.FileServer(http.Dir(staticRoot))
	}
	if sfs != nil {
		mux.Handle("/static/", http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setWasmEncodingHeaders(w, r.URL.Path)
			sfs.ServeHTTP(w, r)
		})))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if sfs != nil {
			spath := filepath.Join(staticRoot, r.URL.Path)
			if st, err := os.Stat(spath); err == nil && !st.IsDir() {
				setWasmEncodingHeaders(w, spath)
				sfs.ServeHTTP(w, r)
				return
			}
		}
		path := filepath.Join(root, r.URL.Path)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			setWasmEncodingHeaders(w, path)
			fs.ServeHTTP(w, r)
			return
		}
		// Serve index.html only for HTML requests or bare paths to avoid
		// returning HTML for CSS, JS, image, etc. requests.
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") || r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		http.NotFound(w, r)
	})
	mux.Handle("/ws", GuardWS(websocket.Handler(wsHandler), opts...))
	return mux
}

// ListenAndServe starts an HTTP server using NewMux to serve files and the
// WebSocket endpoint.
func ListenAndServe(addr, root string) error {
	logger.Info("serving HTTP", "addr", addr)
	return http.ListenAndServe(addr, loggingMiddleware(NewMux(root)))
}

// ListenAndServeWithMux starts an HTTP server using the provided mux.
func ListenAndServeWithMux(addr string, mux *http.ServeMux) error {
	logger.Info("serving HTTP", "addr", addr)
	return http.ListenAndServe(addr, loggingMiddleware(mux))
}

// ListenAndServeTLS starts an HTTPS server using a self-signed certificate
// and NewMux to serve files and the WebSocket endpoint.
func ListenAndServeTLS(addr, root string) error {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:      addr,
		Handler:   loggingMiddleware(NewMux(root)),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
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
	srv := &http.Server{
		Addr:      addr,
		Handler:   loggingMiddleware(mux),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	logger.Info("serving HTTPS", "addr", addr)
	return srv.ListenAndServeTLS("", "")
}

func setWasmEncodingHeaders(w http.ResponseWriter, path string) {
	if !strings.HasSuffix(path, ".wasm.br") {
		return
	}
	header := w.Header()
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
