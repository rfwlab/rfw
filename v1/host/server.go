package host

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

// ListenAndServeTLS starts an HTTPS server using a self-signed certificate
// and NewMux to serve files and the WebSocket endpoint.
func ListenAndServeTLS(addr, root string) error {
	cert, err := generateSelfSignedCert()
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:      addr,
		Handler:   NewMux(root),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
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
