//go:build js && wasm

package events

import (
	js "github.com/rfwlab/rfw/v1/js"
)

// On attaches a handler function for the given event to target.
// Optional opts are forwarded to addEventListener as-is.
// It returns a function that removes the listener and releases resources.
func On(event string, target js.Value, handler func(js.Value), opts ...any) func() {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			handler(args[0])
		} else {
			handler(js.Null())
		}
		return nil
	})
	callArgs := []any{event, fn}
	if len(opts) > 0 {
		callArgs = append(callArgs, opts...)
	}
	target.Call("addEventListener", callArgs...)
	return func() {
		target.Call("removeEventListener", event, fn)
		fn.Release()
	}
}

// OnClick attaches a click handler to target.
func OnClick(target js.Value, handler func(js.Value)) func() {
	return On("click", target, handler)
}

// OnScroll attaches a scroll handler to target.
func OnScroll(target js.Value, handler func(js.Value)) func() {
	return On("scroll", target, handler)
}

// OnInput attaches an input handler to target.
func OnInput(target js.Value, handler func(js.Value)) func() {
	return On("input", target, handler)
}

// OnTimeUpdate attaches a timeupdate handler to target.
func OnTimeUpdate(target js.Value, handler func(js.Value)) func() {
	return On("timeupdate", target, handler)
}

// OnKeyDown attaches a keydown handler to the window object.
func OnKeyDown(handler func(js.Value)) func() {
	return On("keydown", js.Window(), handler)
}

// OnKeyUp attaches a keyup handler to the window object.
func OnKeyUp(handler func(js.Value)) func() {
	return On("keyup", js.Window(), handler)
}

// Listen attaches an event listener to target and returns a channel
// that receives the first argument of the event callback.
func Listen(event string, target js.Value) <-chan js.Value {
	ch := make(chan js.Value)
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			ch <- args[0]
		} else {
			ch <- js.Null()
		}
		return nil
	})
	target.Call("addEventListener", event, fn)
	return ch
}

// ObserveMutations observes DOM mutations on the first node matching sel.
// It returns a channel receiving MutationRecord objects and a stop function
// that disconnects the observer and releases resources.
func ObserveMutations(sel string) (<-chan js.Value, func()) {
	ch := make(chan js.Value)
	node := js.Document().Call("querySelector", sel)
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		mutations := args[0]
		for i := 0; i < mutations.Length(); i++ {
			m := mutations.Index(i)
			t := m.Get("target")
			if t.Truthy() && t.Get("closest").Type() != js.TypeUndefined {
				if t.Call("closest", "[data-rfw-ignore]").Truthy() {
					continue
				}
			}
			ch <- m
		}
		return nil
	})
	observer := js.MutationObserver().New(fn)
	observer.Call("observe", node, js.ValueOf(map[string]any{"childList": true, "subtree": true}))
	stop := func() {
		observer.Call("disconnect")
		fn.Release()
	}
	return ch, stop
}

// ObserveIntersections observes intersections for elements matching sel.
// opts is passed directly to the IntersectionObserver constructor.
// It returns a channel receiving IntersectionObserverEntry objects and a
// stop function to disconnect the observer and release resources.
func ObserveIntersections(sel string, opts js.Value) (<-chan js.Value, func()) {
	ch := make(chan js.Value)
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		entries := args[0]
		for i := 0; i < entries.Length(); i++ {
			ch <- entries.Index(i)
		}
		return nil
	})
	observer := js.IntersectionObserver().New(fn, opts)
	nodes := js.Document().Call("querySelectorAll", sel)
	for i := 0; i < nodes.Length(); i++ {
		observer.Call("observe", nodes.Index(i))
	}
	stop := func() {
		observer.Call("disconnect")
		fn.Release()
	}
	return ch, stop
}
