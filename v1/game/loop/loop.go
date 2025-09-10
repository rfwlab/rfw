//go:build js && wasm

package loop

import (
	"time"

	js "github.com/rfwlab/rfw/v1/js"
)

// Ticker provides timing information for each frame.
type Ticker struct {
	// Delta is the time elapsed since the last frame.
	Delta time.Duration
	// FPS is the calculated frames per second.
	FPS float64
}

var (
	running  bool
	updates  []func(Ticker)
	renders  []func(Ticker)
	handle   js.Func
	lastTime float64

	requestAnimationFrame = js.RequestAnimationFrame
	now                   = func() float64 { return js.Performance().Call("now").Float() }
)

// OnUpdate registers a callback invoked before each frame is rendered.
func OnUpdate(fn func(Ticker)) { updates = append(updates, fn) }

// OnRender registers a callback invoked after updates each frame.
func OnRender(fn func(Ticker)) { renders = append(renders, fn) }

// Start begins the requestAnimationFrame loop.
func Start() {
	if running {
		return
	}
	running = true
	lastTime = now()
	handle = js.FuncOf(func(this js.Value, args []js.Value) any {
		curr := now()
		delta := curr - lastTime
		lastTime = curr
		t := Ticker{Delta: time.Duration(delta * float64(time.Millisecond))}
		if delta > 0 {
			t.FPS = 1000 / delta
		}
		for _, fn := range updates {
			fn(t)
		}
		for _, fn := range renders {
			fn(t)
		}
		if running {
			requestAnimationFrame(handle)
		}
		return nil
	})
	requestAnimationFrame(handle)
}

// Stop halts the animation loop.
func Stop() {
	if !running {
		return
	}
	running = false
	handle.Release()
}
