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
		newValue := args[2].String()

		store := GlobalStoreManager.GetStore(storeName)
		if store == nil {
			store = NewStore(storeName)
		}
		store.Set(key, newValue)
		return nil
	}))
}
