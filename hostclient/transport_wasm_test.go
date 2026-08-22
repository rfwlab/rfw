//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

func TestReadDrainsBufferedFramesBeforeReportingTheClose(t *testing.T) {
	c := newHostConn()
	// Several frames, not one: a single frame is drained by the first
	// non-blocking select and would pass even if the close raced the queue.
	frames := []string{`{"sequence":1}`, `{"sequence":2}`, `{"sequence":3}`}
	for _, frame := range frames {
		c.deliver([]byte(frame))
	}
	c.fail(errors.New("host went away"))

	for i, want := range frames {
		frame, err := c.read(context.Background())
		if err != nil {
			t.Fatalf("read frame %d: %v", i, err)
		}
		if string(frame) != want {
			t.Fatalf("frame %d = %s, want %s", i, frame, want)
		}
	}
	if _, err := c.read(context.Background()); err == nil {
		t.Fatal("read after the queue drained returned no error")
	}
	// The failure must stay readable: a reconnect asks again.
	if _, err := c.read(context.Background()); err == nil {
		t.Fatal("second read after close returned no error")
	}
}

// Liveness counts frames as the browser delivers them, not as readLoop drains
// them, so a slow handler cannot look like a dead connection.
func TestDeliveryCountsArrivalNotConsumption(t *testing.T) {
	c := newHostConn()
	c.deliver([]byte("{}"))
	c.deliver([]byte("{}"))
	if got := c.received.Load(); got != 2 {
		t.Fatalf("received frames = %d, want 2", got)
	}
	if _, err := c.read(context.Background()); err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if got := c.received.Load(); got != 2 {
		t.Fatalf("received frames after a read = %d, want 2", got)
	}
}

func TestInboundQueueOverflowClosesTheConnection(t *testing.T) {
	c := newHostConn()
	for i := 0; i < inboundFrames; i++ {
		c.deliver([]byte("{}"))
	}
	c.deliver([]byte("{}"))

	select {
	case <-c.done:
		if c.err() == nil {
			t.Fatal("overflow reported a nil error")
		}
	default:
		t.Fatal("overflow left the connection open")
	}
}

func TestHeartbeatFailsWithoutInboundTraffic(t *testing.T) {
	c := newHostConn()
	c.send = func(string) error { return nil }

	if err := c.heartbeat(context.Background(), 5*time.Millisecond, 20*time.Millisecond); err == nil {
		t.Fatal("heartbeat accepted a silent connection")
	}
}

func TestHeartbeatSurvivesAnsweredProbes(t *testing.T) {
	c := newHostConn()
	c.send = func(string) error {
		c.received.Add(1)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	if err := c.heartbeat(ctx, 5*time.Millisecond, 10*time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("heartbeat error = %v, want a context deadline", err)
	}
}

func TestControlFrameIsUnsequencedAndNotRetained(t *testing.T) {
	deliveryMu.Lock()
	savedNext := nextOutbound
	savedOutbox := outbox
	savedInbound := lastInbound
	nextOutbound = 7
	outbox = map[uint64]message{}
	lastInbound = 4
	deliveryMu.Unlock()
	defer func() {
		deliveryMu.Lock()
		nextOutbound = savedNext
		outbox = savedOutbox
		lastInbound = savedInbound
		deliveryMu.Unlock()
	}()

	frames := make(chan string, 1)
	c := newHostConn()
	c.send = func(data string) error {
		frames <- data
		return nil
	}
	if err := sendControl(c, "ping"); err != nil {
		t.Fatalf("send control frame: %v", err)
	}

	var frame wireMessage
	if err := json.Unmarshal([]byte(<-frames), &frame); err != nil {
		t.Fatalf("decode control frame: %v", err)
	}
	if frame.Control != "ping" {
		t.Fatalf("control = %q, want ping", frame.Control)
	}
	if frame.Sequence != 0 {
		t.Fatalf("control sequence = %d, want 0", frame.Sequence)
	}
	if frame.Ack != 4 {
		t.Fatalf("control ack = %d, want 4", frame.Ack)
	}

	deliveryMu.Lock()
	retained := len(outbox)
	next := nextOutbound
	deliveryMu.Unlock()
	if retained != 0 {
		t.Fatalf("control frame entered the outbox: %d retained", retained)
	}
	if next != 7 {
		t.Fatalf("control frame consumed a sequence: nextOutbound = %d, want 7", next)
	}
}

// The endpoint is derived from window.location unless the app configured one.
// A page served over https must not downgrade its socket to ws.
func TestNormalizeWSURLDerivesTheScheme(t *testing.T) {
	cases := map[string]string{
		"ws://host.invalid/ws":    "ws://host.invalid/ws",
		"wss://host.invalid/ws":   "wss://host.invalid/ws",
		"http://host.invalid/ws":  "ws://host.invalid/ws",
		"https://host.invalid/ws": "wss://host.invalid/ws",
		"http://host.invalid":     "ws://host.invalid/ws",
		"https://host.invalid":    "wss://host.invalid/ws",
		"https://host.invalid/":   "wss://host.invalid/ws",
		"https://host.invalid//":  "wss://host.invalid/ws",
		"wss://host.invalid/live": "wss://host.invalid/live",
		"  ws://host.invalid/ws ": "ws://host.invalid/ws",
		"":                        "",
	}
	for raw, want := range cases {
		if got := normalizeWSURL(raw); got != want {
			t.Errorf("normalizeWSURL(%q) = %q, want %q", raw, got, want)
		}
	}
}

// A bare host takes the scheme of the page. wasmbrowsertest serves over plain
// HTTP, so the socket must come out as ws and carry the default path.
func TestNormalizeWSURLFollowsThePageProtocol(t *testing.T) {
	got := normalizeWSURL("host.invalid:8080")
	protocol := js.Location().Get("protocol").String()
	want := "wss://host.invalid:8080/ws"
	if protocol == "http:" {
		want = "ws://host.invalid:8080/ws"
	}
	if got != want {
		t.Fatalf("normalizeWSURL on a %s page = %q, want %q", protocol, got, want)
	}
}
