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
	"time"

	"golang.org/x/net/websocket"
)

func resolveRoot(root string) string {
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
// WebSocket handler at /ws.
func NewMux(root string) *http.ServeMux {
	root = resolveRoot(root)
	staticRoot := filepath.Join(root, "..", "static")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(root))
	var sfs http.Handler
	if _, err := os.Stat(staticRoot); err == nil {
		sfs = http.FileServer(http.Dir(staticRoot))
	}
	if sfs != nil {
		mux.Handle("/static/", http.StripPrefix("/static", sfs))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if sfs != nil {
			spath := filepath.Join(staticRoot, r.URL.Path)
			if st, err := os.Stat(spath); err == nil && !st.IsDir() {
				sfs.ServeHTTP(w, r)
				return
			}
		}
		path := filepath.Join(root, r.URL.Path)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
	mux.Handle("/ws", websocket.Handler(wsHandler))
	if os.Getenv("RFW_DEVTOOLS") != "" {
		mux.Handle("/debug/vars", expvar.Handler())
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
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
