//go:build js && wasm

package js

import (
	jst "syscall/js"
)

// Global returns the JavaScript global object.
func Global() jst.Value {
	return jst.Global()
}

// Get returns a property from the global object.
func Get(name string) jst.Value {
	return Global().Get(name)
}

// Set assigns a value on the global object.
func Set(name string, v any) {
	Global().Set(name, v)
}

// Call invokes a function on the global object.
func Call(name string, args ...any) jst.Value {
	return Global().Call(name, args...)
}

// Window returns the window object.
func Window() jst.Value {
	return Get("window")
}

// Win is an alias for Window.
func Win() jst.Value {
	return Window()
}

// Document returns the document object.
func Document() jst.Value {
	return Get("document")
}

// Doc is an alias for Document.
func Doc() jst.Value {
	return Document()
}

// Console returns the console object.
func Console() jst.Value {
	return Get("console")
}

// History returns the history object.
func History() jst.Value {
	return Window().Get("history")
}

// Location returns the location object.
func Location() jst.Value {
	return Window().Get("location")
}

// Loc is an alias for Location.
func Loc() jst.Value {
	return Location()
}

// JSON returns the JSON object.
func JSON() jst.Value {
	return Get("JSON")
}

// Error returns the Error constructor.
func Error() jst.Value {
	return Get("Error")
}

// Performance returns the performance object.
func Performance() jst.Value {
	return Get("performance")
}

// Perf is an alias for Performance.
func Perf() jst.Value {
	return Performance()
}

// MutationObserver returns the MutationObserver constructor.
func MutationObserver() jst.Value {
	return Get("MutationObserver")
}

// IntersectionObserver returns the IntersectionObserver constructor.
func IntersectionObserver() jst.Value {
	return Get("IntersectionObserver")
}

// CustomEvent returns the CustomEvent constructor.
func CustomEvent() jst.Value {
	return Get("CustomEvent")
}

// LocalStorage returns the localStorage object.
func LocalStorage() jst.Value {
	return Get("localStorage")
}

// LS is an alias for LocalStorage.
func LS() jst.Value {
	return LocalStorage()
}

// RequestAnimationFrame wraps the requestAnimationFrame global function.
func RequestAnimationFrame(cb jst.Func) {
	Call("requestAnimationFrame", cb)
}

// RAF is an alias for RequestAnimationFrame.
func RAF(cb jst.Func) {
	RequestAnimationFrame(cb)
}

// Fetch wraps the global fetch function.
func Fetch(args ...any) jst.Value {
	return Call("fetch", args...)
}

// Expose registers a no-argument Go function under the given name
// on the JavaScript global object.
func Expose(name string, fn func()) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) any {
		fn()
		return nil
	}))
}

// ExposeEvent registers a Go function that receives the first argument
// from the JavaScript call as the event object.
func ExposeEvent(name string, fn func(jst.Value)) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) any {
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
func ExposeFunc(name string, fn func(this jst.Value, args []jst.Value) any) {
	Global().Set(name, jst.FuncOf(fn))
}

// Stack returns the current JavaScript stack trace using Error().stack.
func Stack() string {
	return Global().Get("Error").New().Get("stack").String()
}
