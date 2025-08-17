//go:build js && wasm

package events

import (
	jst "syscall/js"
)

// Listen attaches an event listener to target and returns a channel
// that receives the first argument of the event callback.
func Listen(event string, target jst.Value) <-chan jst.Value {
	ch := make(chan jst.Value)
	fn := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		if len(args) > 0 {
			ch <- args[0]
		} else {
			ch <- jst.Null()
		}
		return nil
	})
	target.Call("addEventListener", event, fn)
	return ch
}

// ObserveMutations observes DOM mutations on the first node matching sel.
// It returns a channel receiving MutationRecord objects and a stop function
// that disconnects the observer and releases resources.
func ObserveMutations(sel string) (<-chan jst.Value, func()) {
	ch := make(chan jst.Value)
	document := jst.Global().Get("document")
	node := document.Call("querySelector", sel)
	fn := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		mutations := args[0]
		for i := 0; i < mutations.Length(); i++ {
			ch <- mutations.Index(i)
		}
		return nil
	})
	observer := jst.Global().Get("MutationObserver").New(fn)
	observer.Call("observe", node, jst.ValueOf(map[string]any{"childList": true, "subtree": true}))
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
func ObserveIntersections(sel string, opts jst.Value) (<-chan jst.Value, func()) {
	ch := make(chan jst.Value)
	fn := jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		entries := args[0]
		for i := 0; i < entries.Length(); i++ {
			ch <- entries.Index(i)
		}
		return nil
	})
	observer := jst.Global().Get("IntersectionObserver").New(fn, opts)
	document := jst.Global().Get("document")
	nodes := document.Call("querySelectorAll", sel)
	for i := 0; i < nodes.Length(); i++ {
		observer.Call("observe", nodes.Index(i))
	}
	stop := func() {
		observer.Call("disconnect")
		fn.Release()
	}
	return ch, stop
}
