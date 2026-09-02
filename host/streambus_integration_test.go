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
	mux := NewMux(root)
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
	server := &wt.Server{H3: h3}
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
	response, session, err := dialer.Dial(ctx, url, nil)
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
	inbound, err := json.Marshal(Inbound{Component: "missing", Sequence: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFrame(stream, inbound); err != nil {
		t.Fatal(err)
	}
	payload, err := readFrame(bufio.NewReader(stream), 1<<20)
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
