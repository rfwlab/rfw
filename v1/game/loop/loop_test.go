//go:build js && wasm

package loop

import (
	"math"
	"testing"
	"time"

	js "github.com/rfwlab/rfw/v1/js"
)

func TestLoopCallbacks(t *testing.T) {
	origRAF := requestAnimationFrame
	origNow := now
	origUpdates := updates
	origRenders := renders
	defer func() {
		requestAnimationFrame = origRAF
		now = origNow
		updates = origUpdates
		renders = origRenders
		running = false
	}()

	var frame js.Func
	requestAnimationFrame = func(f js.Func) { frame = f }

	times := []float64{0, 16, 32}
	idx := 0
	now = func() float64 {
		v := times[idx]
		idx++
		return v
	}

	var u, r int
	var tick Ticker
	OnUpdate(func(t Ticker) { u++; tick = t })
	OnRender(func(t Ticker) { r++ })

	Start()

	frame.Invoke()
	if u != 1 || r != 1 {
		t.Fatalf("expected 1 update and render, got %d and %d", u, r)
	}
	if tick.Delta != 16*time.Millisecond {
		t.Fatalf("expected delta 16ms, got %v", tick.Delta)
	}
	if math.Abs(tick.FPS-62.5) > 0.1 {
		t.Fatalf("expected fps about 62.5, got %v", tick.FPS)
	}

	frame.Invoke()
	if u != 2 || r != 2 {
		t.Fatalf("expected 2 updates and renders, got %d and %d", u, r)
	}

	Stop()
}
