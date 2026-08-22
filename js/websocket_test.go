//go:build js && wasm

package js

import "testing"

func TestOpenSocketReportsAConstructorFailure(t *testing.T) {
	for _, url := range []string{"ftp://example.test/ws", "ws://example.test/ws#fragment"} {
		socket, err := OpenSocket(url, SocketHandlers{})
		if err == nil {
			_ = socket.Close(1000, "test")
			t.Fatalf("OpenSocket accepted %q", url)
		}
	}
}

func TestOpenSocketReleasesItsCallbacks(t *testing.T) {
	socket, err := OpenSocket("ws://127.0.0.1:1/ws", SocketHandlers{})
	if err != nil {
		t.Fatalf("open socket: %v", err)
	}
	if len(socket.callbacks) != len(socketEvents) {
		t.Fatalf("bound callbacks = %d, want %d", len(socket.callbacks), len(socketEvents))
	}
	if err := socket.Close(1000, "test"); err != nil {
		t.Fatalf("close socket: %v", err)
	}
	if len(socket.callbacks) != 0 {
		t.Fatalf("callbacks remained bound after close")
	}
	for _, event := range socketEvents {
		if socket.value.Get(event).Truthy() {
			t.Fatalf("%s remained attached after close", event)
		}
	}
	_ = socket.Close(1000, "test")
}

func TestSendTextOnAClosedSocketFails(t *testing.T) {
	socket, err := OpenSocket("ws://127.0.0.1:1/ws", SocketHandlers{})
	if err != nil {
		t.Fatalf("open socket: %v", err)
	}
	defer func() { _ = socket.Close(1000, "test") }()
	if err := socket.Close(1000, "test"); err != nil {
		t.Fatalf("close socket: %v", err)
	}
	if err := socket.SendText("{}"); err != ErrSocketNotOpen {
		t.Fatalf("SendText error = %v, want %v", err, ErrSocketNotOpen)
	}
}

func TestOversizeMessageIsReportedInsteadOfDelivered(t *testing.T) {
	var reported error
	handlers := SocketHandlers{
		MaxMessageBytes: 4,
		Error:           func(err error) { reported = err },
	}
	if !handlers.oversize(5) {
		t.Fatal("payload above the limit was accepted")
	}
	if reported == nil {
		t.Fatal("oversize payload was not reported")
	}
	reported = nil
	if handlers.oversize(4) {
		t.Fatal("payload at the limit was rejected")
	}
	if reported != nil {
		t.Fatalf("payload at the limit reported %v", reported)
	}
	unlimited := SocketHandlers{Error: func(err error) { reported = err }}
	if unlimited.oversize(1 << 30) {
		t.Fatal("an unset limit rejected a payload")
	}
}

// A text frame must reach the Text handler. The first version of this file
// measured the payload with data.Get("length"), which panics on a JavaScript
// string primitive; SafeFuncOf recovered the panic, so every inbound text
// frame was dropped without an error reaching the caller.
func TestTextFrameReachesTheHandler(t *testing.T) {
	Call("eval", `window.__socketProbe = { onmessage: null };`)
	probe := Get("__socketProbe")
	defer Global().Delete("__socketProbe")

	var got string
	var errs []error
	socket := &Socket{value: probe}
	handlers := SocketHandlers{
		MaxMessageBytes: 1 << 20,
		Text:            func(text string) { got = text },
		Error:           func(err error) { errs = append(errs, err) },
	}
	socket.bind("onmessage", func(event Value) { socket.deliver(event, handlers) })
	defer socket.Release()

	event := NewDict()
	event.Set("data", `{"component":"ticker"}`)
	probe.Call("onmessage", event.Value)

	if got != `{"component":"ticker"}` {
		t.Fatalf("Text handler received %q", got)
	}
	if len(errs) != 0 {
		t.Fatalf("delivery reported %v", errs)
	}
}

// An oversize text frame is reported, not delivered.
func TestOversizeTextFrameIsRejected(t *testing.T) {
	Call("eval", `window.__socketProbe2 = { onmessage: null };`)
	probe := Get("__socketProbe2")
	defer Global().Delete("__socketProbe2")

	delivered := false
	var reported error
	socket := &Socket{value: probe}
	handlers := SocketHandlers{
		MaxMessageBytes: 4,
		Text:            func(string) { delivered = true },
		Error:           func(err error) { reported = err },
	}
	socket.bind("onmessage", func(event Value) { socket.deliver(event, handlers) })
	defer socket.Release()

	event := NewDict()
	event.Set("data", "12345")
	probe.Call("onmessage", event.Value)

	if delivered {
		t.Fatal("an oversize frame was delivered")
	}
	if reported == nil {
		t.Fatal("an oversize frame was not reported")
	}
}
