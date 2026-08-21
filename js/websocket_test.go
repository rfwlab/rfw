//go:build js && wasm

package js

import "testing"

func TestOpenSocketReportsAConstructorFailure(t *testing.T) {
	for _, url := range []string{"ftp://example.test/ws", "ws://example.test/ws#fragment"} {
		socket, err := OpenSocket(url, SocketHandlers{})
		if err == nil {
			socket.Close(1000, "test")
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
	socket.Close(1000, "test")
}

func TestSendTextOnAClosedSocketFails(t *testing.T) {
	socket, err := OpenSocket("ws://127.0.0.1:1/ws", SocketHandlers{})
	if err != nil {
		t.Fatalf("open socket: %v", err)
	}
	defer socket.Close(1000, "test")
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
