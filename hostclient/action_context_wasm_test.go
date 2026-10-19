//go:build js && wasm

package hostclient

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCallRefusesCancelledContextBeforeStartingDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := callSequence.Load()
	response, err := Call[map[string]int, int](ctx, "test.cancelled", map[string]int{"delta": 9})
	if !errors.Is(err, context.Canceled) || response != 0 {
		t.Fatalf("cancelled call: response=%d error=%v", response, err)
	}
	if after := callSequence.Load(); after != before {
		t.Fatalf("cancelled call allocated delivery id: before=%d after=%d", before, after)
	}
}

func TestCallRefusesExpiredDeadlineBeforeStartingDelivery(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer cancel()
	before := callSequence.Load()
	response, err := Call[map[string]int, int](ctx, "test.expired", map[string]int{"delta": 9})
	if !errors.Is(err, context.DeadlineExceeded) || response != 0 {
		t.Fatalf("expired call: response=%d error=%v", response, err)
	}
	if after := callSequence.Load(); after != before {
		t.Fatalf("expired call allocated delivery id: before=%d after=%d", before, after)
	}
}

func TestCallRechecksCancellationAfterWaitingForSessionReset(t *testing.T) {
	lifecycleMu.Lock()
	locked := true
	defer func() {
		if locked {
			lifecycleMu.Unlock()
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	before := callSequence.Load()
	done := make(chan error, 1)
	go func() {
		_, err := Call[map[string]int, int](ctx, "test.blocked", map[string]int{"delta": 9})
		done <- err
	}()
	// Observe the blocked call before cancelling; a sleep alone could cancel
	// before the goroutine starts and only exercise the entry check again.
	blocked := false
	stack := make([]byte, 64<<10)
	for i := 0; i < 200; i++ {
		n := runtime.Stack(stack, true)
		if strings.Contains(string(stack[:n]), "hostclient.Call[") {
			blocked = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !blocked {
		t.Fatal("call never reached the locked session lifecycle")
	}
	cancel()
	lifecycleMu.Unlock()
	locked = false
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("blocked call error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled call did not return after lifecycle unlock")
	}
	if after := callSequence.Load(); after != before {
		t.Fatalf("blocked cancelled call allocated delivery id: before=%d after=%d", before, after)
	}
}
