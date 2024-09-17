//go:build js && wasm

package framework

import (
	"fmt"
	"regexp"
	"strings"
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
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		value := input.Get("value").String()
		storeMatch := regexp.MustCompile(`@store:(\w+)\.(\w+):w`).FindStringSubmatch(value)

		if len(storeMatch) == 3 {
			storeName := storeMatch[1]
			key := storeMatch[2]

			store := GlobalStoreManager.GetStore(storeName)
			if store != nil {
				storeValue := store.Get(key)
				if storeValue == nil {
					storeValue = ""
				}
				input.Set("value", fmt.Sprintf("%v", storeValue))

				input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					newValue := this.Get("value").String()
					store.Set(key, newValue)
					return nil
				}))
			}
		}
	}
}

func getConditionDependencies(condition string) ([]ConditionDependency, error) {
	conditionParts := strings.Split(condition, "==")
	if len(conditionParts) != 2 {
		return nil, fmt.Errorf("Invalid condition format")
	}

	leftSide := strings.TrimSpace(conditionParts[0])
	leftSide = strings.Replace(leftSide, "@if:", "", 1)

	dependencies := []ConditionDependency{}

	if strings.HasPrefix(leftSide, "store:") {
		storeParts := strings.Split(strings.TrimPrefix(leftSide, "store:"), ".")
		if len(storeParts) == 2 {
			storeName, key := storeParts[0], storeParts[1]
			dependencies = append(dependencies, ConditionDependency{storeName, key})
		}
	}

	return dependencies, nil
}
