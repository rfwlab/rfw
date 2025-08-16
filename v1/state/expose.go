package state

import "syscall/js"

// ExposeUpdateStore exposes a JS function to update store values.
func ExposeUpdateStore() {
	js.Global().Set("goUpdateStore", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return nil
		}
		storeName := args[0].String()
		key := args[1].String()

		var newValue interface{}
		switch args[2].Type() {
		case js.TypeString:
			newValue = args[2].String()
		case js.TypeBoolean:
			newValue = args[2].Bool()
		case js.TypeNumber:
			newValue = args[2].Float()
		default:
			newValue = args[2]
		}

		store := GlobalStoreManager.GetStore(storeName)
		if store == nil {
			store = NewStore(storeName)
		}
		store.Set(key, newValue)
		return nil
	}))
}
