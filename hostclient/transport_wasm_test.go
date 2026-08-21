//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestReadDrainsBufferedFramesBeforeReportingTheClose(t *testing.T) {
	c := newHostConn()
	c.deliver([]byte(`{"control":"ack"}`))
	c.fail(errors.New("host went away"))

	frame, err := c.read(context.Background())
	if err != nil {
		t.Fatalf("read buffered frame: %v", err)
	}
	if string(frame) != `{"control":"ack"}` {
		t.Fatalf("buffered frame = %s", frame)
	}
	if _, err := c.read(context.Background()); err == nil {
		t.Fatal("read after close returned no error")
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
	case err := <-c.closed:
		if err == nil {
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
