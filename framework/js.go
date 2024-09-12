//go:build js && wasm

package framework

import (
	"syscall/js"
)

func ExposeFunction(name string, fn interface{}) {
	js.Global().Set(name, js.FuncOf(fn.(func(this js.Value, args []js.Value) interface{})))
}
