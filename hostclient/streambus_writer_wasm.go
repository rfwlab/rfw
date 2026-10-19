//go:build js && wasm

package hostclient

import (
	"context"
	"errors"
	"sync"
)

const (
	streamBusOutboundFrames = 256
	streamBusOutboundBytes  = 8 << 20
)

// Browser callbacks cannot wait on stream promises. One worker writes whole
// frames in order; the budget includes the frame waiting on the native writer.
func startStreamBusWriter(c *hostConn, writer *jsWriter) {
	queue := make(chan string, streamBusOutboundFrames)
	var mu sync.Mutex
	var frames, bytes int
	c.send = func(data string) error {
		mu.Lock()
		defer mu.Unlock()
		select {
		case <-c.done:
			return c.err()
		default:
		}
		if frames == streamBusOutboundFrames || len(data) > streamBusOutboundBytes-bytes {
			err := errors.New("hostclient: StreamBus outbound queue overflow")
			c.fail(err)
			return err
		}
		frames++
		bytes += len(data)
		queue <- data
		return nil
	}
	go func() {
		defer func() {
			if recover() != nil {
				c.fail(errors.New("hostclient: StreamBus write failed"))
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			select {
			case <-c.done:
				cancel()
			case <-ctx.Done():
			}
		}()
		for {
			select {
			case <-c.done:
				return
			case data := <-queue:
				select {
				case <-c.done:
					return
				default:
				}
				if err := writeStreamBusFrame(ctx, writer, []byte(data)); err != nil {
					c.fail(err)
					return
				}
				mu.Lock()
				frames--
				bytes -= len(data)
				mu.Unlock()
			}
		}
	}()
}
