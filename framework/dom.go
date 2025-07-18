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

	patchInnerHTML(element, html)

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

func patchInnerHTML(element js.Value, html string) {
	document := js.Global().Get("document")
	template := document.Call("createElement", "template")
	template.Set("innerHTML", html)
	newContent := template.Get("content")
	patchChildren(element, newContent)
}

func patchChildren(oldParent, newParent js.Value) {
	oldChildren := oldParent.Get("childNodes")
	newChildren := newParent.Get("childNodes")
	oldLen := oldChildren.Length()
	newLen := newChildren.Length()

	minLen := oldLen
	if newLen < minLen {
		minLen = newLen
	}

	for i := 0; i < minLen; i++ {
		patchNode(oldChildren.Index(i), newChildren.Index(i))
	}

	for i := minLen; i < newLen; i++ {
		clone := newChildren.Index(i).Call("cloneNode", true)
		oldParent.Call("appendChild", clone)
	}

	for i := oldLen - 1; i >= minLen; i-- {
		oldChildren.Index(i).Call("remove")
	}
}

func patchNode(oldNode, newNode js.Value) {
	nodeType := newNode.Get("nodeType").Int()
	if nodeType == 3 { // Text node
		if oldNode.Get("nodeValue").String() != newNode.Get("nodeValue").String() {
			oldNode.Set("nodeValue", newNode.Get("nodeValue"))
		}
		return
	}

	if oldNode.Get("nodeName").String() != newNode.Get("nodeName").String() {
		oldNode.Call("replaceWith", newNode.Call("cloneNode", true))
		return
	}

	patchAttributes(oldNode, newNode)
	patchChildren(oldNode, newNode)
}

func patchAttributes(oldNode, newNode js.Value) {
	oldAttrs := oldNode.Call("getAttributeNames")
	for i := 0; i < oldAttrs.Length(); i++ {
		name := oldAttrs.Index(i).String()
		if !newNode.Call("hasAttribute", name).Bool() {
			oldNode.Call("removeAttribute", name)
		}
	}

	newAttrs := newNode.Call("getAttributeNames")
	for i := 0; i < newAttrs.Length(); i++ {
		name := newAttrs.Index(i).String()
		val := newNode.Call("getAttribute", name)
		if oldNode.Call("getAttribute", name).String() != val.String() {
			oldNode.Call("setAttribute", name, val)
		}
	}
}
