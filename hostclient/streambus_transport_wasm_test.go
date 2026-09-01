//go:build js && wasm

package hostclient

import (
	stdjs "syscall/js"
	"testing"
)

func TestStreamBusURLUsesAdvertisedHTTP3Port(t *testing.T) {
	got := replaceURLPort("https://localhost:8081/streambus", "8083")
	if got != "https://localhost:8083/streambus" {
		t.Fatalf("URL = %q", got)
	}
}

func TestNormalizeStreamBusURLFromDevelopmentWebSocket(t *testing.T) {
	got := normalizeStreamBusURL("ws://localhost:8080/ws", true)
	if got != "https://localhost:8081/streambus" {
		t.Fatalf("URL = %q", got)
	}
}

func TestPreferredHostTransportReadsGeneratedConfig(t *testing.T) {
	global := stdjs.Global()
	previous := global.Get("RFW_TRANSPORT")
	global.Set("RFW_TRANSPORT", "streambus")
	t.Cleanup(func() {
		if previous.Type() == stdjs.TypeUndefined {
			global.Delete("RFW_TRANSPORT")
		} else {
			global.Set("RFW_TRANSPORT", previous)
		}
	})
	if got := preferredHostTransport(); got != hostTransportStreamBus {
		t.Fatalf("transport = %q", got)
	}
}
