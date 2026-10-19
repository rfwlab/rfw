package host

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	wt "github.com/quic-go/webtransport-go"
)

func TestStreamBusWebTransportProtocol(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	testStreamBusMux(t, NewMux(root))
}

func testStreamBusMux(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	testStreamBusMuxCredential(t, mux, "", false)
}

func TestStreamBusCredentialRestoresCookieIdentityAndRefusesReuse(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	var initialized atomic.Int64
	mux := NewMux(t.TempDir(), WithSSCStreamBusAuth(), WithAuthFunc(func(r *http.Request) bool {
		cookie, err := r.Cookie("session")
		return err == nil && cookie.Value == "test-user"
	}), WithSSCSessionInitializer(func(r *http.Request, _ *Session) error {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "test-user" {
			return fmt.Errorf("identity missing")
		}
		initialized.Add(1)
		return nil
	}))
	endpoint, ok := streamEndpoints.Load(mux)
	if !ok {
		t.Fatal("missing StreamBus endpoint")
	}
	token := issueStreamBusTestCredential(t, endpoint.(*streamBusEndpoint).runtime.Guard(http.HandlerFunc(endpoint.(*streamBusEndpoint).issueCredential)))
	testStreamBusMuxCredential(t, mux, token, false)
	testStreamBusMuxCredential(t, mux, token, true)
	if initialized.Load() != 1 {
		t.Fatalf("authenticated sessions initialized = %d", initialized.Load())
	}
}

func TestStreamBusUnknownCredentialCannotInitializeSession(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	var initialized atomic.Int64
	mux := NewMux(t.TempDir(), WithSSCStreamBusAuth(), WithAuthFunc(func(*http.Request) bool { return true }),
		WithSSCSessionInitializer(func(*http.Request, *Session) error { initialized.Add(1); return nil }))
	testStreamBusMuxCredential(t, mux, strings.Repeat("0", 64), true)
	if initialized.Load() != 0 {
		t.Fatalf("unauthorized sessions initialized = %d", initialized.Load())
	}
}

func testStreamBusMuxCredential(t *testing.T, mux *http.ServeMux, token string, rejected bool) {
	t.Helper()
	t.Cleanup(func() { streamEndpoints.Delete(mux) })
	value, ok := streamEndpoints.Load(mux)
	if !ok {
		t.Fatal("StreamBus endpoint was not registered")
	}
	endpoint := value.(*streamBusEndpoint)

	certificate, err := generateSelfSignedCert()
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(leaf)
	serverTLS := &tls.Config{Certificates: []tls.Certificate{certificate}, NextProtos: []string{http3.NextProtoH3}}
	h3 := &http3.Server{
		TLSConfig: serverTLS, Handler: mux, EnableDatagrams: true,
		QUICConfig: &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
	}
	// The production server delegates origin policy to the mux guard.
	server := &wt.Server{H3: h3, CheckOrigin: func(*http.Request) bool { return true }}
	wt.ConfigureHTTP3Server(h3)
	endpoint.mu.Lock()
	endpoint.server = server
	endpoint.mu.Unlock()

	address, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	packet, err := net.ListenUDP("udp", address)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(packet) }()
	t.Cleanup(func() {
		_ = server.Close()
		_ = packet.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("WebTransport server did not stop")
		}
	})

	dialer := &wt.Dialer{
		TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: "localhost"},
		QUICConfig:      &quic.Config{EnableDatagrams: true, EnableStreamResetPartialDelivery: true},
	}
	t.Cleanup(func() { _ = dialer.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://localhost:%d%s", packet.LocalAddr().(*net.UDPAddr).Port, streamBusPath)
	headers := http.Header{}
	if token != "" {
		headers.Set("Origin", "https://app.test")
	}
	response, session, err := dialer.Dial(ctx, url, headers)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	t.Cleanup(func() { _ = session.CloseWithError(0, "") })
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(stream)
	if token != "" {
		payload, err := json.Marshal(map[string]string{"token": token})
		if err != nil {
			t.Fatal(err)
		}
		if err := writeFrame(stream, payload); err != nil {
			t.Fatal(err)
		}
		ack, err := readFrame(reader, 4096)
		if rejected {
			if err == nil {
				t.Fatalf("invalid credential received %d acknowledgement bytes", len(ack))
			}
			return
		}
		if err != nil || string(ack) != `{"control":"authenticated"}` {
			t.Fatalf("transport authentication: error=%v ack bytes=%d", err, len(ack))
		}
	}
	inbound, err := json.Marshal(Inbound{Component: "missing", Sequence: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFrame(stream, inbound); err != nil {
		t.Fatal(err)
	}
	payload, err := readFrame(reader, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var outbound Outbound
	if err := json.Unmarshal(payload, &outbound); err != nil {
		t.Fatal(err)
	}
	if outbound.Control != "ack" || outbound.Session == "" || outbound.Sequence == 0 {
		t.Fatalf("unexpected response: %#v", outbound)
	}
}
