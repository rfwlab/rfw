package core

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestScopeCancelsAndCleansUpOnce(t *testing.T) {
	scope := NewScope()
	var cleanups atomic.Int32
	cancelled := make(chan struct{})
	scope.Defer(func() { cleanups.Add(1) })
	scope.Go(func(ctx context.Context) {
		<-ctx.Done()
		close(cancelled)
	})

	scope.Close()
	scope.Close()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("scope context was not cancelled")
	}
	if cleanups.Load() != 1 {
		t.Fatalf("cleanup count = %d", cleanups.Load())
	}
}

func TestScopeRunsLateCleanupImmediately(t *testing.T) {
	scope := NewScope()
	scope.Close()
	called := false
	scope.Defer(func() { called = true })
	if !called {
		t.Fatal("late cleanup did not run")
	}
}

func TestScopeRunsRemainingCleanupAfterPanic(t *testing.T) {
	scope := NewScope()
	ran := false
	scope.Defer(func() { ran = true })
	scope.Defer(func() { panic("cleanup") })
	scope.Close()
	if !ran {
		t.Fatal("cleanup after panic did not run")
	}
}
