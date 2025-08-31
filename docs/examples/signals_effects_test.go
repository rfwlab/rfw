package main

import (
	"testing"

	state "github.com/rfwlab/rfw/v1/state"
)

func TestSignalAndEffect(t *testing.T) {
	count := state.NewSignal(0)
	runs := 0
	stop := state.Effect(func() func() {
		_ = count.Get()
		runs++
		return nil
	})
	count.Set(1)
	if runs != 2 {
		t.Fatalf("expected effect to run twice, got %d", runs)
	}
	stop()
}
