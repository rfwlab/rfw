//go:build js && wasm

package framework

import (
	"syscall/js"
)

func ExposeFunction(name string, fn func()) {
	js.Global().Set(name, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fn()
		return nil
	}))
}
