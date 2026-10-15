//go:build !js || !wasm

package host

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

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
	defer func() { _ = packets.Close() }()
	defer func() { _ = server.Close() }()
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
	defer func() { _ = listener.Close() }()
	packets, err := net.ListenPacket("udp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = packets.Close() }()
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
