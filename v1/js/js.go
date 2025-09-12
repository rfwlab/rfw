//go:build js && wasm

package js

import (
	jst "syscall/js"
)

// Type aliases to re-export syscall/js types.
type (
	Value = jst.Value
	Func  = jst.Func
	Type  = jst.Type
)

// Re-exported Value type constants.
const (
	TypeUndefined = jst.TypeUndefined
	TypeNull      = jst.TypeNull
	TypeBoolean   = jst.TypeBoolean
	TypeNumber    = jst.TypeNumber
	TypeString    = jst.TypeString
	TypeSymbol    = jst.TypeSymbol
	TypeObject    = jst.TypeObject
	TypeFunction  = jst.TypeFunction
)

// Global returns the JavaScript global object.
func Global() Value { return jst.Global() }

// Get returns a property from the global object.
func Get(name string) Value { return Global().Get(name) }

// Set assigns a value on the global object.
func Set(name string, v any) { Global().Set(name, v) }

// Call invokes a function on the global object.
func Call(name string, args ...any) Value { return Global().Call(name, args...) }

// Null returns the JavaScript null value.
func Null() Value { return jst.Null() }

// Undefined returns the JavaScript undefined value.
func Undefined() Value { return jst.Undefined() }

// ValueOf converts a Go value to a JavaScript value.
func ValueOf(v any) Value { return jst.ValueOf(v) }

// Array represents a JavaScript Array instance.
type Array struct{ Value }

// NewArray constructs a new JavaScript Array.
func NewArray(items ...any) Array {
	return Array{Value: Get("Array").New(items...)}
}

// ArrayOf wraps an existing JavaScript value as an Array.
func ArrayOf(v Value) Array { return Array{Value: v} }

// Push appends items to the array and returns the new length.
func (a Array) Push(items ...any) int {
	return a.Call("push", items...).Int()
}

// Concat merges the array with additional arrays.
func (a Array) Concat(arrs ...Array) Array {
	args := make([]any, len(arrs))
	for i, arr := range arrs {
		args[i] = arr.Value
	}
	return Array{Value: a.Call("concat", args...)}
}

// Length reports the number of elements in the array.
func (a Array) Length() int { return a.Get("length").Int() }

// TypedArrayOf converts a Go slice to a JavaScript typed array.
func TypedArrayOf(slice any) jst.Value {
	switch b := slice.(type) {
	case []byte:
		u8 := Uint8Array().New(len(b))
		jst.CopyBytesToJS(u8, b)
		return u8

	case []float32:
		f32 := Float32Array().New(len(b))
		for i, v := range b {
			f32.SetIndex(i, v)
		}
		return f32

	case []uint16:
		u16 := Uint16Array().New(len(b))
		for i, v := range b {
			u16.SetIndex(i, v)
		}
		return u16
	}

	return jst.ValueOf(slice)
}

// FuncOf wraps a Go function for use in JavaScript.
func FuncOf(fn func(this Value, args []Value) any) Func { return jst.FuncOf(fn) }

// Window returns the window object.
func Window() Value { return Get("window") }

// Win is an alias for Window.
func Win() Value { return Window() }

// Document returns the document object.
func Document() Value { return Get("document") }

// Doc is an alias for Document.
func Doc() Value { return Document() }

// Console returns the console object.
func Console() Value { return Get("console") }

// History returns the history object.
func History() Value { return Window().Get("history") }

// Location returns the location object.
func Location() Value { return Window().Get("location") }

// Loc is an alias for Location.
func Loc() Value { return Location() }

// JSON returns the JSON object.
func JSON() Value { return Get("JSON") }

// Error returns the Error constructor.
func Error() Value { return Get("Error") }

// Performance returns the performance object.
func Performance() Value { return Get("performance") }

// Perf is an alias for Performance.
func Perf() Value { return Performance() }

// MutationObserver returns the MutationObserver constructor.
func MutationObserver() Value { return Get("MutationObserver") }

// IntersectionObserver returns the IntersectionObserver constructor.
func IntersectionObserver() Value { return Get("IntersectionObserver") }

// CustomEvent returns the CustomEvent constructor.
func CustomEvent() Value { return Get("CustomEvent") }

// Uint8Array returns the Uint8Array constructor.
func Uint8Array() Value { return Get("Uint8Array") }

// Float32Array returns the Float32Array constructor.
func Float32Array() Value { return Get("Float32Array") }

// Uint16Array returns the Uint16Array constructor.
func Uint16Array() Value { return Get("Uint16Array") }

// LocalStorage returns the localStorage object.
func LocalStorage() Value { return Get("localStorage") }

// LS is an alias for LocalStorage.
func LS() Value { return LocalStorage() }

// RequestAnimationFrame wraps the requestAnimationFrame global function.
func RequestAnimationFrame(cb Func) { Call("requestAnimationFrame", cb) }

// RAF is an alias for RequestAnimationFrame.
func RAF(cb Func) { RequestAnimationFrame(cb) }

// Fetch wraps the global fetch function.
func Fetch(args ...any) Value { return Call("fetch", args...) }

// Expose registers a no-argument Go function under the given name
// on the JavaScript global object.
func Expose(name string, fn func()) {
	Global().Set(name, FuncOf(func(this Value, args []Value) any {
		fn()
		return nil
	}))
}

// ExposeEvent registers a Go function that receives the first argument
// from the JavaScript call as the event object.
func ExposeEvent(name string, fn func(Value)) {
	Global().Set(name, FuncOf(func(this Value, args []Value) any {
		var evt Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	}))
}

// ExposeFunc registers a Go function with custom arguments on the
// JavaScript global object.
func ExposeFunc(name string, fn func(this Value, args []Value) any) {
	Global().Set(name, FuncOf(fn))
}

// Stack returns the current JavaScript stack trace using Error().stack.
func Stack() string { return Global().Get("Error").New().Get("stack").String() }
