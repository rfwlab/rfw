//go:build js && wasm

package js

import "syscall/js"

// Expose registers a no-argument Go function under the given name
// on the JavaScript global object.
func Expose(name string, fn func()) {
	js.Global().Set(name, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fn()
		return nil
	}))
}

// ExposeEvent registers a Go function that receives the first argument
// from the JavaScript call as the event object.
func ExposeEvent(name string, fn func(js.Value)) {
	js.Global().Set(name, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var evt js.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	}))
}
