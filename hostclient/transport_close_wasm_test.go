//go:build js && wasm

package hostclient

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLocalCloseEndsReadsWithoutRemoteAcknowledgement(t *testing.T) {
	c := newHostConn()
	c.shutdown = func() error { return nil }
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	read := make(chan error, 1)
	go func() {
		_, err := c.read(ctx)
		read <- err
	}()
	if err := c.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := <-read; !errors.Is(err, errConnectionClosed) {
		t.Fatalf("read after local close: %v", err)
	}
}

func TestLocalCloseEndsLogicalConnectionBeforeNativeShutdown(t *testing.T) {
	c := newHostConn()
	c.shutdown = func() error {
		select {
		case <-c.done:
		default:
			t.Error("native shutdown started with the logical connection open")
		}
		return nil
	}
	if err := c.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestLocalClosePreservesFailureAndReturnsShutdownError(t *testing.T) {
	c := newHostConn()
	failure := errors.New("transport failed")
	shutdownErr := errors.New("native shutdown failed")
	c.fail(failure)
	c.shutdown = func() error { return shutdownErr }
	if err := c.close(); !errors.Is(err, shutdownErr) {
		t.Fatalf("close: %v", err)
	}
	if _, err := c.read(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("read after close: %v", err)
	}
}
