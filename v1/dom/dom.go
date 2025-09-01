//go:build js && wasm

// Package dom provides utilities for updating the browser DOM and binding
// event listeners for WebAssembly components.
package dom

import (
	"fmt"
	"regexp"
	"strings"
	jst "syscall/js"

	events "github.com/rfwlab/rfw/v1/events"
	"github.com/rfwlab/rfw/v1/state"
)

// componentSignals tracks signals associated with each component instance.
var componentSignals = make(map[string]map[string]any)

// RegisterSignal associates a signal with a component so inputs can bind to it.
func RegisterSignal(componentID, name string, sig any) {
	if componentSignals[componentID] == nil {
		componentSignals[componentID] = make(map[string]any)
	}
	componentSignals[componentID][name] = sig
}

// RemoveComponentSignals cleans up signals for a component on unmount.
func RemoveComponentSignals(componentID string) { delete(componentSignals, componentID) }

func getSignal(componentID, name string) any {
	if m, ok := componentSignals[componentID]; ok {
		return m[name]
	}
	return nil
}

// TemplateHook is an optional callback invoked after a DOM update to allow
// custom processing of the rendered HTML.
var TemplateHook func(componentID, html string)

// UpdateDOM patches the DOM of the specified component with the provided
// HTML string.
func UpdateDOM(componentID string, html string) {
	var element jst.Value
	if componentID == "" {
		element = ByID("app")
	} else {
		element = Query(fmt.Sprintf("[data-component-id='%s']", componentID))
		if element.IsNull() || element.IsUndefined() {
			element = ByID("app")
		}
	}
	if element.IsNull() || element.IsUndefined() {
		return
	}

	RemoveEventListeners(componentID)

	if strings.HasPrefix(html, "<root") && strings.EqualFold(element.Get("nodeName").String(), "ROOT") {
		if end := strings.LastIndex(html, "</root>"); end != -1 {
			if start := strings.Index(html, ">"); start != -1 {
				html = html[start+1 : end]
			}
		}
	}

	patchInnerHTML(element, html)

	if TemplateHook != nil {
		TemplateHook(componentID, html)
	}

	BindStoreInputs(element)
	BindSignalInputs(componentID, element)

	BindEventListeners(componentID, element)
}

// BindStoreInputs binds input elements to store variables.
func BindStoreInputs(element jst.Value) {
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)

		valueAttr := input.Get("value").String()
		checkedAttr := ""
		if input.Call("hasAttribute", "checked").Bool() {
			checkedAttr = input.Call("getAttribute", "checked").String()
		}

		re := regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+):w`)
		valueMatch := re.FindStringSubmatch(valueAttr)
		checkedMatch := re.FindStringSubmatch(checkedAttr)

		var module, storeName, key string
		var usesChecked bool
		if len(valueMatch) == 4 {
			module, storeName, key = valueMatch[1], valueMatch[2], valueMatch[3]
		} else if len(checkedMatch) == 4 {
			module, storeName, key = checkedMatch[1], checkedMatch[2], checkedMatch[3]
			usesChecked = true
		} else {
			continue
		}

		store := state.GlobalStoreManager.GetStore(module, storeName)
		if store == nil {
			continue
		}
		storeValue := store.Get(key)

		if usesChecked {
			boolVal, _ := storeValue.(bool)
			input.Set("checked", boolVal)
			ch := events.Listen("change", input)
			go func(in jst.Value, st *state.Store, k string) {
				for range ch {
					st.Set(k, in.Get("checked").Bool())
				}
			}(input, store, key)
			continue
		}

		if storeValue == nil {
			storeValue = ""
		}
		input.Set("value", fmt.Sprintf("%v", storeValue))
		ch := events.Listen("input", input)
		go func(in jst.Value, st *state.Store, k string) {
			for range ch {
				st.Set(k, in.Get("value").String())
			}
		}(input, store, key)
	}
}

// BindSignalInputs binds input elements to local component signals.
func BindSignalInputs(componentID string, element jst.Value) {
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)

		valueAttr := input.Get("value").String()
		checkedAttr := ""
		if input.Call("hasAttribute", "checked").Bool() {
			checkedAttr = input.Call("getAttribute", "checked").String()
		}

		re := regexp.MustCompile(`@signal:(\w+):w`)
		valueMatch := re.FindStringSubmatch(valueAttr)
		checkedMatch := re.FindStringSubmatch(checkedAttr)

		var name string
		var usesChecked bool
		if len(valueMatch) == 2 {
			name = valueMatch[1]
		} else if len(checkedMatch) == 2 {
			name = checkedMatch[1]
			usesChecked = true
		} else {
			continue
		}

		sig := getSignal(componentID, name)
		if sig == nil {
			continue
		}

		if usesChecked {
			if s, ok := sig.(interface {
				Read() any
				Set(bool)
			}); ok {
				if b, ok := s.Read().(bool); ok {
					input.Set("checked", b)
				}
				ch := events.Listen("change", input)
				go func(in jst.Value, sg interface{ Set(bool) }) {
					for range ch {
						sg.Set(in.Get("checked").Bool())
					}
				}(input, s)
			}
			continue
		}

		if s, ok := sig.(interface {
			Read() any
			Set(string)
		}); ok {
			input.Set("value", fmt.Sprintf("%v", s.Read()))
			ch := events.Listen("input", input)
			go func(in jst.Value, sg interface{ Set(string) }) {
				for range ch {
					sg.Set(in.Get("value").String())
				}
			}(input, s)
		}
	}
}

func patchInnerHTML(element jst.Value, html string) {
	template := CreateElement("template")
	template.Set("innerHTML", html)
	newContent := template.Get("content")
	patchChildren(element, newContent)
}

func patchChildren(oldParent, newParent jst.Value) {
	oldChildren := oldParent.Get("childNodes")
	newChildren := newParent.Get("childNodes")

	keyed := make(map[string]jst.Value)
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

func getDataKey(node jst.Value) string {
	if node.Get("nodeType").Int() != 1 {
		return ""
	}
	key := node.Call("getAttribute", "data-key")
	if key.Truthy() {
		return key.String()
	}
	return ""
}

func patchNode(oldNode, newNode jst.Value) {
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

func patchAttributes(oldNode, newNode jst.Value) {
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
