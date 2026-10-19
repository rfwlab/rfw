//go:build js && wasm

package hostclient

import (
	"context"
	"strings"
	"testing"

	js "github.com/rfwlab/rfw/v2/js"
)

func installStreamBusSetup(t *testing.T, constructor string) {
	t.Helper()
	window := js.Window()
	previous := map[string]js.Value{
		"WebTransport":   window.Get("WebTransport"),
		"fetch":          window.Get("fetch"),
		"rfwSetupCloses": window.Get("rfwSetupCloses"),
	}
	t.Cleanup(func() {
		for key, value := range previous {
			if value.IsUndefined() {
				window.Delete(key)
			} else {
				window.Set(key, value)
			}
		}
	})
	window.Set("rfwSetupCloses", 0)
	window.Set("fetch", window.Get("Function").New("return Promise.resolve({ok:false})"))
	window.Set("WebTransport", window.Get("Function").New("return "+constructor).Invoke())
}

// A policy or URL error can throw before ready exists; the caller must still
// get an error that allows its documented WebSocket fallback.
func TestStreamBusConstructorExceptionReturnsAnError(t *testing.T) {
	installStreamBusSetup(t, `class { constructor() { throw new Error('policy refused') } }`)
	connection, err := dialStreamBus(context.Background(), "https://localhost/streambus")
	if connection != nil || err == nil || !strings.Contains(err.Error(), "WebTransport setup") {
		t.Fatalf("constructor result: connection=%v error=%v", connection, err)
	}
}

func TestStreamBusStreamSetupExceptionClosesTheTransport(t *testing.T) {
	installStreamBusSetup(t, `class {
		constructor() { this.ready = Promise.resolve() }
		createBidirectionalStream() { throw new Error('stream refused') }
		close() { window.rfwSetupCloses++ }
	}`)
	connection, err := dialStreamBus(context.Background(), "https://localhost/streambus")
	if connection != nil || err == nil || js.Window().Get("rfwSetupCloses").Int() != 1 {
		t.Fatalf("stream setup: connection=%v error=%v closes=%d", connection, err, js.Window().Get("rfwSetupCloses").Int())
	}
}

func TestStreamBusRejectedReadinessClosesTheTransportOnce(t *testing.T) {
	installStreamBusSetup(t, `class {
		constructor() { this.ready = Promise.reject(new Error('handshake refused')) }
		close() { window.rfwSetupCloses++ }
	}`)
	connection, err := dialStreamBus(context.Background(), "https://localhost/streambus")
	if connection != nil || err == nil || js.Window().Get("rfwSetupCloses").Int() != 1 {
		t.Fatalf("readiness: connection=%v error=%v closes=%d", connection, err, js.Window().Get("rfwSetupCloses").Int())
	}
}

func TestStreamBusCleanupExceptionPreservesTheSetupError(t *testing.T) {
	installStreamBusSetup(t, `class {
		constructor() { this.ready = Promise.resolve() }
		createBidirectionalStream() { throw new Error('stream refused') }
		close() { window.rfwSetupCloses++; throw new Error('close refused') }
	}`)
	connection, err := dialStreamBus(context.Background(), "https://localhost/streambus")
	if connection != nil || err == nil || !strings.Contains(err.Error(), "WebTransport setup") || js.Window().Get("rfwSetupCloses").Int() != 1 {
		t.Fatalf("cleanup: connection=%v error=%v closes=%d", connection, err, js.Window().Get("rfwSetupCloses").Int())
	}
}
