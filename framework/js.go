//go:build js && wasm

package framework

import (
	"syscall/js"
)

func ExposeFunction(name string, fn interface{}) {
	js.Global().Set(name, js.FuncOf(fn.(func(this js.Value, args []js.Value) interface{})))
}

func ExposeUpdateStore() {
	js.Global().Set("goUpdateStore", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return nil
		}
		storeName := args[0].String()
		key := args[1].String()
		newValue := args[2].String()

		store := GlobalStoreManager.GetStore(storeName)
		if store == nil {
			store = NewStore(storeName)
		}
		store.Set(key, newValue)
		return nil
	}))
}
