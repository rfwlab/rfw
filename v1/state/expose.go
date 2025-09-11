//go:build js && wasm

package state

import (
	js "github.com/rfwlab/rfw/v1/js"
)

// ExposeUpdateStore exposes a JS function to update store values.
func ExposeUpdateStore() {
	js.Set("goUpdateStore", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 4 {
			return nil
		}
		module := args[0].String()
		storeName := args[1].String()
		key := args[2].String()

		var newValue any
		switch args[3].Type() {
		case js.TypeString:
			newValue = args[3].String()
		case js.TypeBoolean:
			newValue = args[3].Bool()
		case js.TypeNumber:
			newValue = args[3].Float()
		default:
			newValue = args[3]
		}

		store := GlobalStoreManager.GetStore(module, storeName)
		if store == nil {
			store = NewStore(storeName, WithModule(module))
		}
		store.Set(key, newValue)
		return nil
	}))
}
