//go:build js && wasm

package http

import (
	"testing"
	"time"
)

func TestRegisterHTTPHook(t *testing.T) {
	var starts, completes int
	RegisterHTTPHook(func(start bool, _ string, _ int, _ time.Duration) {
		if start {
			starts++
		} else {
			completes++
		}
	})
	if httpHook == nil {
		t.Fatal("hook not registered")
	}
	httpHook(true, "u", 0, 0)
	httpHook(false, "u", 200, time.Millisecond)
	if starts != 1 || completes != 1 {
		t.Fatalf("expected 1 start and 1 complete, got %d and %d", starts, completes)
	}
}
