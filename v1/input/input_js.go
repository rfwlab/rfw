//go:build js && wasm

package input

import (
	"github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

// New creates a Manager and wires browser event listeners.
func New() *Manager {
	m := newManager()

	events.OnKeyDown(func(evt js.Value) {
		m.handleKeyDown(evt.Get("key").String())
	})
	events.OnKeyUp(func(evt js.Value) {
		m.handleKeyUp(evt.Get("key").String())
	})

	events.On("mousedown", js.Document(), func(evt js.Value) {
		m.handleMouseDown(evt.Get("button").Int(), float32(evt.Get("clientX").Float()), float32(evt.Get("clientY").Float()))
	})
	events.On("mouseup", js.Document(), func(evt js.Value) {
		m.handleMouseUp(evt.Get("button").Int(), float32(evt.Get("clientX").Float()), float32(evt.Get("clientY").Float()))
	})
	events.On("mousemove", js.Document(), func(evt js.Value) {
		m.handleMouseMove(float32(evt.Get("clientX").Float()), float32(evt.Get("clientY").Float()))
	})
	events.On("wheel", js.Document(), func(evt js.Value) {
		m.handleWheel(float32(evt.Get("deltaY").Float()))
	})

	return m
}
