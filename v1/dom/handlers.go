//go:build js && wasm

package dom

import (
	js "github.com/rfwlab/rfw/v1/js"
)

var handlerRegistry = make(map[string]js.Value)

// RegisterHandler registers a Go function with custom arguments in the handler registry.
func RegisterHandler(name string, fn func(this js.Value, args []js.Value) any) {
	handlerRegistry[name] = js.FuncOf(fn).Value
}

// RegisterHandlerFunc registers a no-argument Go function in the handler registry.
func RegisterHandlerFunc(name string, fn func()) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		fn()
		return nil
	})
}

// RegisterHandlerEvent registers a Go function that receives the first argument as an event object.
func RegisterHandlerEvent(name string, fn func(js.Value)) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		var evt js.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	})
}

// GetHandler retrieves a registered handler by name.
func GetHandler(name string) js.Value {
	if v, ok := handlerRegistry[name]; ok {
		return v
	}
	return js.Undefined()
}
