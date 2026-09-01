//go:build js && wasm

package hostclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

const (
	heartbeatInterval = 30 * time.Second
	heartbeatTimeout  = 5 * time.Second

	// inboundFrames buffers frames between the JavaScript event loop and
	// readLoop. The browser delivers a message on a callback that must not
	// block, so an overflow drops the connection instead of stalling the page.
	inboundFrames = 256

	// normalClosure is the only close code the browser WebSocket API accepts
	// outside the 3000-4999 application range.
	normalClosure = 1000
)

var errConnectionClosed = errors.New("hostclient: connection closed")

// hostConn adapts the browser WebSocket to the blocking read and write calls
// the connection loops expect.
type hostConn struct {
	socket   *js.Socket
	send     func(string) error
	shutdown func() error
	release  func()
	// generation binds the transport to the authenticated client session that
	// created it. ResetSession advances the global generation before closing
	// the socket, so a late frame from the old native/browser callback cannot
	// reach handlers owned by the replacement identity.
	generation uint64
	messages   chan []byte
	open       chan struct{}
	// done is closed once, so every reader observes the end of the connection.
	// A one-shot error channel would hand the failure to whichever read
	// happened to be waiting and leave the next one blocked forever.
	done chan struct{}

	openOnce  sync.Once
	closeOnce sync.Once
	closeErr  atomic.Pointer[error]
	received  atomic.Uint64
}

func newHostConn() *hostConn {
	return &hostConn{
		messages: make(chan []byte, inboundFrames),
		open:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// dialBrowser opens a connection and waits for the browser handshake to
// complete. It remains the default SSC transport for existing applications.
func dialBrowser(ctx context.Context, url string) (*hostConn, error) {
	c := newHostConn()
	socket, err := js.OpenSocket(url, js.SocketHandlers{
		MaxMessageBytes: int(maxInboundMessageBytes),
		Open:            func() { c.openOnce.Do(func() { close(c.open) }) },
		Text:            func(text string) { c.deliver([]byte(text)) },
		Bytes:           c.deliver,
		Error:           c.fail,
		Close: func(code int, reason string, _ bool) {
			c.fail(fmt.Errorf("hostclient: connection closed with code %d: %s", code, reason))
		},
	})
	if err != nil {
		return nil, err
	}
	c.socket = socket
	c.send = socket.SendText
	c.shutdown = func() error { return socket.Close(normalClosure, "connection closed") }

	select {
	case <-c.open:
		return c, nil
	case <-c.done:
		_ = c.close()
		return nil, c.err()
	case <-ctx.Done():
		_ = c.close()
		return nil, ctx.Err()
	}
}

func (c *hostConn) deliver(payload []byte) {
	if int64(len(payload)) > maxInboundMessageBytes {
		c.fail(fmt.Errorf(
			"hostclient: inbound message of %d bytes exceeds the %d byte limit",
			len(payload), maxInboundMessageBytes,
		))
		return
	}
	select {
	case c.messages <- payload:
		c.received.Add(1)
	default:
		c.fail(errors.New("hostclient: inbound frame queue overflow"))
	}
}

func (c *hostConn) fail(err error) {
	c.closeOnce.Do(func() {
		if err == nil {
			err = errConnectionClosed
		}
		c.closeErr.Store(&err)
		close(c.done)
	})
}

// err reports why the connection ended, once done is closed.
func (c *hostConn) err() error {
	if stored := c.closeErr.Load(); stored != nil {
		return *stored
	}
	return errConnectionClosed
}

// read returns the next frame. Buffered frames always win over a close: a
// select with both cases ready picks at random, so the queue is drained
// without blocking first and only an empty queue lets the close through. A
// connection that ends mid-burst still delivers every frame the host sent,
// which is what ordered delivery depends on.
func (c *hostConn) read(ctx context.Context) ([]byte, error) {
	select {
	case payload := <-c.messages:
		return payload, nil
	default:
	}
	select {
	case payload := <-c.messages:
		return payload, nil
	case <-c.done:
		// The close may have raced a frame that landed between the two
		// selects. done stays closed, so returning the frame now still lets
		// the next read report the failure.
		select {
		case payload := <-c.messages:
			return payload, nil
		default:
		}
		return nil, c.err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *hostConn) writeJSON(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if c.send == nil {
		return errConnectionClosed
	}
	return c.send(string(data))
}

func (c *hostConn) close() error {
	var err error
	if c.shutdown != nil {
		err = c.shutdown()
	} else if c.socket != nil {
		err = c.socket.Close(normalClosure, "connection closed")
	}
	if c.release != nil {
		c.release()
		c.release = nil
	}
	return err
}

// heartbeat probes liveness with a control message. The browser WebSocket API
// gives script no access to protocol ping and pong frames, so a half-open
// connection can only be detected by asking the host to answer over the same
// JSON channel.
func (c *hostConn) heartbeat(ctx context.Context, interval, timeout time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
		// Snapshot before probing. deliver counts a frame when the browser
		// hands it over, not when readLoop drains it, so every frame already
		// queued is behind this number: a half-open socket cannot hide behind
		// a backlog. Any arrival counts, not just the pong, which is what lets
		// a host that predates the control message answer with its generic ack.
		delivered := c.received.Load()
		if err := sendControl(c, "ping"); err != nil {
			return err
		}
		timer := time.NewTimer(timeout)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
		if c.received.Load() == delivered {
			return errors.New("hostclient: heartbeat timed out")
		}
	}
}

// sendControl writes an unsequenced control frame. It carries no sequence so it
// never enters the outbox and is never replayed after a reconnect.
func sendControl(c *hostConn, control string) error {
	deliveryMu.Lock()
	ack := lastInbound
	deliveryMu.Unlock()
	sendMu.Lock()
	defer sendMu.Unlock()
	return c.writeJSON(wireMessage{Control: control, Ack: ack})
}
