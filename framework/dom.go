// Path: /home/mirko/Projects/personal/rfw/framework/dom.go
//go:build js && wasm

package framework

import (
	"fmt"
	"syscall/js"
)

func UpdateDOM(componentID string, html string) {
	document := js.Global().Get("document")
	var element js.Value
	if componentID == "" {
		element = document.Call("getElementById", "app")
	} else {
		element = document.Call("querySelector", fmt.Sprintf("[data-component-id='%s']", componentID))
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}
	element.Set("innerHTML", html)

	bindStoreInputs(element)
}

func bindStoreInputs(element js.Value) {
	inputs := element.Call("querySelectorAll", "input[data-store][data-key]")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		storeName := input.Get("dataset").Get("store").String()
		key := input.Get("dataset").Get("key").String()

		input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			newValue := this.Get("value").String()
			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				store.Set(key, newValue)
			}
			return nil
		}))
	}
}
