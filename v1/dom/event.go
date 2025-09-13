//go:build js && wasm

package dom

import js "github.com/rfwlab/rfw/v1/js"

// Event wraps a browser event.
type Event struct{ js.Value }

// On attaches a listener for event to the element and returns a stop function.
func (e Element) On(event string, handler func(Event)) func() {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		var evt js.Value
		if len(args) > 0 {
			evt = args[0]
		}
		handler(Event{evt})
		return nil
	})
	e.Call("addEventListener", event, fn)
	return func() {
		e.Call("removeEventListener", event, fn)
		fn.Release()
	}
}

// OnClick attaches a click handler to the element.
func (e Element) OnClick(handler func(Event)) func() {
	return e.On("click", handler)
}

// PreventDefault prevents the default action for the event.
func (e Event) PreventDefault() { e.Call("preventDefault") }

// StopPropagation stops the event from bubbling.
func (e Event) StopPropagation() { e.Call("stopPropagation") }
