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
	messages chan []byte
	open     chan struct{}
	closed   chan error

	openOnce  sync.Once
	closeOnce sync.Once
	received  atomic.Uint64
}

func newHostConn() *hostConn {
	return &hostConn{
		messages: make(chan []byte, inboundFrames),
		open:     make(chan struct{}),
		closed:   make(chan error, 1),
	}
}

// dial opens a connection and waits for the browser handshake to complete.
func dial(ctx context.Context, url string) (*hostConn, error) {
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

	select {
	case <-c.open:
		return c, nil
	case err := <-c.closed:
		c.close()
		return nil, err
	case <-ctx.Done():
		c.close()
		return nil, ctx.Err()
	}
}

func (c *hostConn) deliver(payload []byte) {
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
		c.closed <- err
	})
}

// read returns the next frame, preferring buffered frames over a close so a
// connection that ends mid-burst still delivers what the host already sent.
func (c *hostConn) read(ctx context.Context) ([]byte, error) {
	select {
	case payload := <-c.messages:
		return payload, nil
	default:
	}
	select {
	case payload := <-c.messages:
		return payload, nil
	case err := <-c.closed:
		return nil, err
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
	if c.socket == nil {
		return nil
	}
	return c.socket.Close(normalClosure, "connection closed")
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
