//go:build js && wasm

package dom

import (
	jst "syscall/js"
)

var handlerRegistry = make(map[string]jst.Value)

// RegisterHandler registers a Go function with custom arguments in the handler registry.
func RegisterHandler(name string, fn func(this jst.Value, args []jst.Value) interface{}) {
	handlerRegistry[name] = jst.FuncOf(fn).Value
}

// RegisterHandlerFunc registers a no-argument Go function in the handler registry.
func RegisterHandlerFunc(name string, fn func()) {
	RegisterHandler(name, func(this jst.Value, args []jst.Value) interface{} {
		fn()
		return nil
	})
}

// RegisterHandlerEvent registers a Go function that receives the first argument as an event object.
func RegisterHandlerEvent(name string, fn func(jst.Value)) {
	RegisterHandler(name, func(this jst.Value, args []jst.Value) interface{} {
		var evt jst.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	})
}

// GetHandler retrieves a registered handler by name.
func GetHandler(name string) jst.Value {
	if v, ok := handlerRegistry[name]; ok {
		return v
	}
	return jst.Undefined()
}
