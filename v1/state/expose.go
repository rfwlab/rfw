//go:build js && wasm

package state

import (
	jst "syscall/js"

	js "github.com/rfwlab/rfw/v1/js"
)

// ExposeUpdateStore exposes a JS function to update store values.
func ExposeUpdateStore() {
	js.Set("goUpdateStore", jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		if len(args) < 4 {
			return nil
		}
		module := args[0].String()
		storeName := args[1].String()
		key := args[2].String()

		var newValue interface{}
		switch args[3].Type() {
		case jst.TypeString:
			newValue = args[3].String()
		case jst.TypeBoolean:
			newValue = args[3].Bool()
		case jst.TypeNumber:
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
