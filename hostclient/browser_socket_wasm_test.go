//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

// The browser gives script no way to run a WebSocket server, so these tests
// drive the real transport against a scripted constructor installed on window.
// Everything below the fake is production code: js.OpenSocket, hostConn and
// readLoop all run exactly as they do against a live host.
const fakeSocketSource = `
(function () {
  window.__fakeSockets = [];
  window.__FakeWebSocket = function (url) {
    var self = this;
    this.url = url;
    this.readyState = 0;
    this.sent = [];
    this.binaryType = "blob";
    this.closeCalls = [];
    this.send = function (data) { self.sent.push(data); };
    this.close = function (code, reason) {
      self.closeCalls.push({ code: code, reason: reason });
      if (self.readyState === 3) { return; }
      self.readyState = 3;
      if (self.onclose) {
        self.onclose({ code: code || 1000, reason: reason || "", wasClean: true });
      }
    };
    window.__fakeSockets.push(this);
  };
})();
`

type fakeSocket struct {
	t     *testing.T
	value js.Value
}

// installFakeSockets swaps window.WebSocket for the scripted constructor and
// restores the real one on cleanup. Cleanup rather than a deferred call in the
// test body: a failed assertion must not leave the fake installed, or the
// tests in js/websocket_test.go that dial for real start failing instead.
func installFakeSockets(t *testing.T) {
	t.Helper()
	js.Call("eval", fakeSocketSource)
	original := js.Get("WebSocket")
	js.Set("WebSocket", js.Get("__FakeWebSocket"))
	js.Set("__fakeSockets", js.NewArray().Value)
	t.Cleanup(func() {
		js.Set("WebSocket", original)
		js.Global().Delete("__fakeSockets")
		js.Global().Delete("__FakeWebSocket")
	})
}

// lastSocket returns the most recently constructed fake.
func lastSocket(t *testing.T) fakeSocket {
	t.Helper()
	sockets := js.Get("__fakeSockets")
	length := sockets.Get("length").Int()
	if length == 0 {
		t.Fatal("no socket was constructed")
	}
	return fakeSocket{t: t, value: sockets.Index(length - 1)}
}

func socketCount() int { return js.Get("__fakeSockets").Get("length").Int() }

// open moves the socket to OPEN and fires onopen, the way a browser does once
// the handshake completes.
func (f fakeSocket) open() {
	f.value.Set("readyState", js.ValueOf(1))
	f.value.Call("onopen", js.NewDict().Value)
}

// deliver fires onmessage with a text frame.
func (f fakeSocket) deliver(frame string) {
	event := js.NewDict()
	event.Set("data", frame)
	f.value.Call("onmessage", event.Value)
}

// serverClose fires onclose the way a host closing the connection does.
func (f fakeSocket) serverClose(code int, reason string) {
	f.value.Set("readyState", js.ValueOf(3))
	event := js.NewDict()
	event.Set("code", code)
	event.Set("reason", reason)
	event.Set("wasClean", true)
	f.value.Call("onclose", event.Value)
}

// sent returns every frame the client wrote to this socket.
func (f fakeSocket) sent() []string {
	raw := f.value.Get("sent")
	out := make([]string, raw.Get("length").Int())
	for i := range out {
		out[i] = raw.Index(i).String()
	}
	return out
}

// handlersBound reports whether the socket still has Go callbacks attached.
func (f fakeSocket) handlersBound() bool {
	for _, event := range []string{"onopen", "onmessage", "onerror", "onclose"} {
		if f.value.Get(event).Truthy() {
			return true
		}
	}
	return false
}

// dialFake opens a connection through the fake and completes the handshake.
func dialFake(t *testing.T) (*hostConn, fakeSocket) {
	t.Helper()
	installFakeSockets(t)
	resetDeliveryState(t)

	type dialed struct {
		conn *hostConn
		err  error
	}
	done := make(chan dialed, 1)
	go func() {
		c, err := dial(context.Background(), "ws://host.invalid/ws")
		done <- dialed{conn: c, err: err}
	}()

	socket := waitForSocket(t)
	socket.open()

	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("dial: %v", result.err)
		}
		t.Cleanup(func() { _ = result.conn.close() })
		return result.conn, socket
	case <-time.After(2 * time.Second):
		t.Fatal("dial did not complete after the handshake")
		return nil, fakeSocket{}
	}
}

// waitForSocket yields to the JavaScript event loop until the constructor has
// run. The dial goroutine reaches OpenSocket only once it is scheduled.
func waitForSocket(t *testing.T) fakeSocket {
	t.Helper()
	for i := 0; i < 200; i++ {
		if socketCount() > 0 {
			return lastSocket(t)
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the client never constructed a socket")
	return fakeSocket{}
}

// resetDeliveryState clears the package level delivery bookkeeping so tests do
// not inherit sequence numbers or pending calls from each other.
func resetDeliveryState(t *testing.T) {
	t.Helper()
	deliveryMu.Lock()
	lastInbound = 0
	nextOutbound = 0
	resumeToken = ""
	outbox = map[uint64]message{}
	deliveryMu.Unlock()
	callMu.Lock()
	pendingCalls = map[string]chan actionReply{}
	callMu.Unlock()
	t.Cleanup(func() {
		deliveryMu.Lock()
		lastInbound = 0
		nextOutbound = 0
		resumeToken = ""
		outbox = map[uint64]message{}
		deliveryMu.Unlock()
		callMu.Lock()
		pendingCalls = map[string]chan actionReply{}
		callMu.Unlock()
	})
}

// readOnce runs readLoop until it returns, so a test can assert on the error a
// single frame produces.
func readOnce(c *hostConn, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return readLoop(ctx, c)
}

func TestBrowserSocketCompletesTheHandshakeAndWrites(t *testing.T) {
	conn, socket := dialFake(t)

	if got := socket.value.Get("binaryType").String(); got != "arraybuffer" {
		t.Fatalf("binaryType = %q, want arraybuffer", got)
	}
	if got := socket.value.Get("url").String(); got != "ws://host.invalid/ws" {
		t.Fatalf("url = %q", got)
	}
	if err := conn.writeJSON(wireMessage{Component: "counter", Sequence: 1}); err != nil {
		t.Fatalf("write: %v", err)
	}
	frames := socket.sent()
	if len(frames) != 1 || !strings.Contains(frames[0], `"component":"counter"`) {
		t.Fatalf("frames written = %v", frames)
	}
}

func TestBrowserSocketDeliversAnActionReply(t *testing.T) {
	conn, socket := dialFake(t)

	reply := make(chan actionReply, 1)
	callMu.Lock()
	pendingCalls["call-1"] = reply
	callMu.Unlock()

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"id":"call-1","payload":{"total":7},"sequence":1}`)

	select {
	case got := <-reply:
		payload, _ := got.payload.(map[string]any)
		if payload["total"] != float64(7) {
			t.Fatalf("reply payload = %#v", got.payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the action reply never reached its caller")
	}
}

func TestBrowserSocketDeliversASubscriptionPayload(t *testing.T) {
	conn, socket := dialFake(t)

	received := make(chan map[string]any, 1)
	mu.Lock()
	handlers["ticker"] = func(payload map[string]any) { received <- payload }
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(handlers, "ticker")
		mu.Unlock()
	})

	go func() { _ = readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"ticker","payload":{"price":1.5},"sequence":1}`)

	select {
	case payload := <-received:
		if payload["price"] != 1.5 {
			t.Fatalf("handler payload = %#v", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the subscription payload never reached its handler")
	}
}

func TestBrowserSocketRejectsAMalformedFrame(t *testing.T) {
	conn, socket := dialFake(t)

	errCh := make(chan error, 1)
	go func() { errCh <- readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":`)

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("a malformed frame was accepted")
		}
		if _, ok := err.(*json.SyntaxError); !ok {
			t.Fatalf("malformed frame error = %v (%T), want a JSON syntax error", err, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a malformed frame froze the read loop")
	}
}

func TestBrowserSocketReportsAServerClosure(t *testing.T) {
	conn, socket := dialFake(t)

	errCh := make(chan error, 1)
	go func() { errCh <- readOnce(conn, 2*time.Second) }()
	socket.serverClose(1001, "going away")

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("a server closure was not reported")
		}
		if !strings.Contains(err.Error(), "1001") {
			t.Fatalf("closure error = %v, want the close code", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a server closure left the read loop running")
	}
}

// A frame that skips a sequence desyncs the session instead of applying out of
// order, which is what forces the reconnect and replay path.
func TestBrowserSocketRejectsASequenceGap(t *testing.T) {
	conn, socket := dialFake(t)

	errCh := make(chan error, 1)
	go func() { errCh <- readOnce(conn, 2*time.Second) }()
	socket.deliver(`{"component":"ticker","sequence":1}`)
	socket.deliver(`{"component":"ticker","sequence":3}`)

	select {
	case err := <-errCh:
		if err == nil || !strings.Contains(err.Error(), "sequence gap") {
			t.Fatalf("sequence gap error = %v", err)
		}
		if got := connectionState.Get(); got != ConnectionDesynced {
			t.Fatalf("connection state = %v, want desynced", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a sequence gap did not end the read loop")
	}
}

// The host rejects a resume it cannot honour with a control frame. The client
// must consume it without treating it as component traffic.
func TestBrowserSocketConsumesARejectedResume(t *testing.T) {
	conn, socket := dialFake(t)

	mu.Lock()
	handlers[""] = func(map[string]any) { t.Error("a control frame reached a component handler") }
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(handlers, "")
		mu.Unlock()
	})

	errCh := make(chan error, 1)
	go func() { errCh <- readOnce(conn, 400*time.Millisecond) }()
	socket.deliver(`{"control":"resume_rejected","error":{"code":"resume_rejected","message":"session could not be resumed"}}`)

	if err := <-errCh; err != context.DeadlineExceeded {
		t.Fatalf("read loop ended with %v, want it still running until the deadline", err)
	}
}

func TestBrowserSocketReconnectsOnANewSocket(t *testing.T) {
	conn, socket := dialFake(t)
	socket.serverClose(1006, "abnormal")
	if _, err := conn.read(context.Background()); err == nil {
		t.Fatal("the closed connection kept reading")
	}

	before := socketCount()
	type dialed struct {
		conn *hostConn
		err  error
	}
	done := make(chan dialed, 1)
	go func() {
		c, err := dial(context.Background(), "ws://host.invalid/ws")
		done <- dialed{conn: c, err: err}
	}()

	for i := 0; i < 200 && socketCount() == before; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if socketCount() <= before {
		t.Fatal("the reconnect reused the closed socket")
	}
	lastSocket(t).open()

	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("reconnect dial: %v", result.err)
		}
		defer func() { _ = result.conn.close() }()
	case <-time.After(2 * time.Second):
		t.Fatal("the reconnect never completed its handshake")
	}
}

// Closing a connection must leave nothing attached: no Go callbacks on the
// socket and no js.Func values still registered with the runtime.
func TestBrowserSocketReleasesEverythingOnClose(t *testing.T) {
	conn, socket := dialFake(t)
	if !socket.handlersBound() {
		t.Fatal("the socket had no callbacks bound while open")
	}
	if err := conn.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if socket.handlersBound() {
		t.Fatal("callbacks stayed attached after close")
	}
	if state := conn.socket.ReadyState(); state != js.SocketClosed {
		t.Fatalf("socket readyState after close = %d, want closed", state)
	}
	// A second close is what a reconnect after an error does; it must not
	// panic or re-release the Go functions.
	if err := conn.close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
