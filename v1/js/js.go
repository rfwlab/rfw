//go:build js && wasm

package js

import (
	jst "syscall/js"
)

// Global returns the JavaScript global object.
func Global() jst.Value {
	return jst.Global()
}

// Expose registers a no-argument Go function under the given name
// on the JavaScript global object.
func Expose(name string, fn func()) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		fn()
		return nil
	}))
}

// ExposeEvent registers a Go function that receives the first argument
// from the JavaScript call as the event object.
func ExposeEvent(name string, fn func(jst.Value)) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		var evt jst.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	}))
}

// ExposeFunc registers a Go function with custom arguments on the
// JavaScript global object.
func ExposeFunc(name string, fn func(this jst.Value, args []jst.Value) interface{}) {
	Global().Set(name, jst.FuncOf(fn))
}

// Stack returns the current JavaScript stack trace using Error().stack.
func Stack() string {
	return Global().Get("Error").New().Get("stack").String()
}
