//go:build js && wasm

package dom

import (
	"fmt"
	"regexp"
	"syscall/js"

	"github.com/rfwlab/rfw/v1/state"
)

var TemplateHook func(componentID, html string)

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

	RemoveEventListeners(componentID)

	patchInnerHTML(element, html)

	if TemplateHook != nil {
		TemplateHook(componentID, html)
	}

	BindStoreInputs(element)

	BindEventListeners(componentID, element)
}

// BindStoreInputs binds input elements to store variables.
func BindStoreInputs(element js.Value) {
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		value := input.Get("value").String()
		storeMatch := regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+):w`).FindStringSubmatch(value)

		if len(storeMatch) == 4 {
			module := storeMatch[1]
			storeName := storeMatch[2]
			key := storeMatch[3]

			store := state.GlobalStoreManager.GetStore(module, storeName)
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

	keyed := make(map[string]js.Value)
	for i := 0; i < oldChildren.Length(); i++ {
		child := oldChildren.Index(i)
		if key := getDataKey(child); key != "" {
			keyed[key] = child
		}
	}

	index := 0
	for i := 0; i < newChildren.Length(); i++ {
		newChild := newChildren.Index(i)
		key := getDataKey(newChild)
		if key != "" {
			if oldChild, ok := keyed[key]; ok {
				patchNode(oldChild, newChild)
				ref := oldParent.Get("childNodes").Index(index)
				if !oldChild.Equal(ref) {
					if ref.Truthy() {
						oldParent.Call("insertBefore", oldChild, ref)
					} else {
						oldParent.Call("appendChild", oldChild)
					}
				}
				delete(keyed, key)
			} else {
				clone := newChild.Call("cloneNode", true)
				ref := oldParent.Get("childNodes").Index(index)
				if ref.Truthy() {
					oldParent.Call("insertBefore", clone, ref)
				} else {
					oldParent.Call("appendChild", clone)
				}
			}
			index++
			continue
		}

		oldChild := oldParent.Get("childNodes").Index(index)
		if oldChild.Truthy() && getDataKey(oldChild) == "" {
			patchNode(oldChild, newChild)
		} else {
			clone := newChild.Call("cloneNode", true)
			ref := oldParent.Get("childNodes").Index(index)
			if ref.Truthy() {
				oldParent.Call("insertBefore", clone, ref)
			} else {
				oldParent.Call("appendChild", clone)
			}
		}
		index++
	}

	for _, child := range keyed {
		child.Call("remove")
	}

	for oldParent.Get("childNodes").Length() > index {
		oldParent.Get("childNodes").Index(index).Call("remove")
	}
}

func getDataKey(node js.Value) string {
	if node.Get("nodeType").Int() != 1 {
		return ""
	}
	key := node.Call("getAttribute", "data-key")
	if key.Truthy() {
		return key.String()
	}
	return ""
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
