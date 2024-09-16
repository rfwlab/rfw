// Path: /home/mirko/Projects/personal/rfw/framework/dom.go
//go:build js && wasm

package framework

import (
	"fmt"
	"regexp"
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
	inputs := element.Call("querySelectorAll", "input")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		storeMatch := regexp.MustCompile(`@store:(\w+)\.(\w+)(:w)?`).FindStringSubmatch(input.Get("value").String())

		if len(storeMatch) >= 3 {
			storeName := storeMatch[1]
			key := storeMatch[2]
			isWriteable := len(storeMatch) == 4 && storeMatch[3] == ":w"

			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				input.Set("value", store.Get(key))

				if isWriteable {
					input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
						newValue := this.Get("value").String()
						store.Set(key, newValue)
						return nil
					}))
				}
			}
		}
	}
}
