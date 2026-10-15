package host

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func TestReadPortOverride(t *testing.T) {
	t.Setenv("RFW_HOST_PORT", "9095")
	if got := readPort(); got != 9095 {
		t.Fatalf("expected override port 9095, got %d", got)
	}
}

func TestStreamBusPreparationAdvertisesHTTP3OnActualPort(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	cert, err := generateSelfSignedCert()
	if err != nil {
		t.Fatal(err)
	}
	server, packets, err := prepareStreamBusHTTP3("127.0.0.1:0", NewMux(t.TempDir()), &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer packets.Close()
	defer server.Close()
	config := server.H3.TLSConfig
	if len(config.NextProtos) != 1 || config.NextProtos[0] != http3.NextProtoH3 {
		t.Fatalf("HTTP/3 protocols are %v", config.NextProtos)
	}
	_, port, err := net.SplitHostPort(packets.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := streamHTTP3Port.Load().(string); got != port {
		t.Fatalf("advertised port is %q, socket port is %q", got, port)
	}
}

func TestHostRefusesOccupiedStreamBusPortBeforeServing(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	packets, err := net.ListenPacket("udp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer packets.Close()
	result := make(chan error, 1)
	mux := NewMux(t.TempDir())
	go func() { result <- serveHost(nil, listener, mux) }()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("host returned no bind error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("host kept running with an unavailable StreamBus port")
	}
}

func TestStreamBusKeepsConfiguredOriginAndIdentityGuards(t *testing.T) {
	t.Setenv("RFW_TRANSPORT", "streambus")
	mux := NewMux(t.TempDir(), WithOriginAllowlist("https://allowed.example"), WithAuthFunc(func(r *http.Request) bool { return r.Header.Get("Authorization") == "valid" }))
	request := httptest.NewRequest("GET", "https://host/streambus", nil)
	request.Header.Set("Origin", "https://denied.example")
	request.Header.Set("Authorization", "valid")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("denied origin returned %d", response.Code)
	}
	request.Header.Set("Origin", "https://allowed.example")
	request.Header.Del("Authorization")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing identity returned %d", response.Code)
	}
	request.Header.Set("Authorization", "valid")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUpgradeRequired {
		t.Fatalf("allowed request returned %d", response.Code)
	}
}
