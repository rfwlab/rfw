//go:build !js || !wasm

package testkit

import (
	"sync/atomic"
	"testing"
	"time"
)

type renderComponent struct{}

func (renderComponent) Render() string  { return "<p>rendered</p>" }
func (renderComponent) GetName() string { return "render" }
func (renderComponent) GetID() string   { return "render" }

func TestRenderAndEventually(t *testing.T) {
	AssertContains(t, Render(renderComponent{}), "rendered")
	var ready atomic.Bool
	go func() {
		time.Sleep(time.Millisecond)
		ready.Store(true)
	}()
	Eventually(t, time.Second, ready.Load)
}
