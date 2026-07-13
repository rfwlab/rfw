//go:build js && wasm

// Package dom provides utilities for updating the browser DOM and binding
// event listeners for WebAssembly components.
package dom

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	events "github.com/rfwlab/rfw/v2/events"
	js "github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
)

// componentSignals tracks signals associated with each component instance.
var (
	componentSignals   = make(map[string]map[string]any)
	componentSignalsMu sync.RWMutex
)

// inputBindingStops tracks the stop functions of input listeners created via
// events.Listen for each component, so rebinding and unmount can release the
// previous listeners instead of leaking them.
var (
	inputBindingStops   = make(map[string][]func())
	inputBindingStopsMu sync.Mutex
)

func addInputBindingStop(componentID string, stop func()) {
	inputBindingStopsMu.Lock()
	inputBindingStops[componentID] = append(inputBindingStops[componentID], stop)
	inputBindingStopsMu.Unlock()
}

// ReleaseInputBindings stops all input listeners registered for a component.
// UpdateDOM calls it before rebinding and core calls it on unmount.
func ReleaseInputBindings(componentID string) {
	inputBindingStopsMu.Lock()
	stops := inputBindingStops[componentID]
	delete(inputBindingStops, componentID)
	inputBindingStopsMu.Unlock()
	for _, stop := range stops {
		stop()
	}
}

var (
	reStoreWrite  = regexp.MustCompile(`@store:(\w+)\.(\w+)\.(\w+):w`)
	reSignalWrite = regexp.MustCompile(`@signal:(\w+):w`)
)

// RegisterSignal associates a signal with a component so inputs can bind to it.
func RegisterSignal(componentID, name string, sig any) {
	componentSignalsMu.Lock()
	if componentSignals[componentID] == nil {
		componentSignals[componentID] = make(map[string]any)
	}
	componentSignals[componentID][name] = sig
	componentSignalsMu.Unlock()
}

// RemoveComponentSignals cleans up signals for a component on unmount.
func RemoveComponentSignals(componentID string) {
	componentSignalsMu.Lock()
	delete(componentSignals, componentID)
	componentSignalsMu.Unlock()
}

func getSignal(componentID, name string) any {
	componentSignalsMu.RLock()
	defer componentSignalsMu.RUnlock()
	if m, ok := componentSignals[componentID]; ok {
		return m[name]
	}
	return nil
}

// SnapshotComponentSignals returns a copy of the signals registered for a component.
func SnapshotComponentSignals(componentID string) map[string]any {
	componentSignalsMu.RLock()
	defer componentSignalsMu.RUnlock()
	if signals, ok := componentSignals[componentID]; ok {
		clone := make(map[string]any, len(signals))
		for k, v := range signals {
			clone[k] = v
		}
		return clone
	}
	return nil
}

// TemplateHook is an optional callback invoked after a DOM update to allow
// custom processing of the rendered HTML.
var TemplateHook func(componentID, html string)

// StoreBindingHook, when set, is invoked for each @store binding associated with
// a component. It receives the component identifier along with the store module,
// store name, and key that are bound in the DOM.
var StoreBindingHook func(componentID, module, store, key string)

// ComponentRoot returns the DOM root element for a component by its ID.
// Falls back to #app if id is empty or element not found.
func ComponentRoot(id string) Element {
	doc := Doc()
	if id == "" {
		return doc.ByID("app")
	}
	el := doc.Query(fmt.Sprintf("[data-component-id='%s']", id))
	if el.IsNull() || el.IsUndefined() {
		return doc.ByID("app")
	}
	return el
}

// UpdateDOM patches the DOM of the specified component with the provided
// HTML string, resolving the target via typed Document/Element wrappers.
func UpdateDOM(componentID string, html string) {
	element := ComponentRoot(componentID)
	if element.IsNull() || element.IsUndefined() {
		return
	}

	// Diff-patch only when the resolved element is the component's OWN root: that
	// is an in-place reactive update, where patching preserves focus/selection.
	// Otherwise the target is the #app fallback (a fresh mount or a route change,
	// since ComponentRoot falls back to #app when the component root is not yet
	// in the DOM). There, positionally diffing two different <root> trees leaves
	// stale nodes from the previous component, so replace wholesale instead.
	elID := element.Value.Call("getAttribute", "data-component-id")
	if componentID != "" && elID.Truthy() && elID.String() == componentID {
		patchInnerHTML(element.Value, html)
	} else {
		element.Value.Set("innerHTML", html)
	}

	if TemplateHook != nil {
		TemplateHook(componentID, html)
	}

	// Release the listeners of the previous render: rebinding below attaches
	// fresh ones and stale listeners on replaced nodes would leak.
	ReleaseInputBindings(componentID)

	BindStoreInputsForComponent(componentID, element.Value)
	BindSignalInputs(componentID, element.Value)
	BindASTStoreInputs(componentID, element.Value)
	BindASTSignalInputs(componentID, element.Value)
}

// UpdateMountedDOM patches a component's subtree only when its own root is in
// the DOM. Reactive updates (store/signal changes) go through here: a change
// hitting a component that is not mounted yet (a constructor-time Set) or not
// anymore must be a no-op, not a wholesale replacement of the #app fallback.
func UpdateMountedDOM(componentID, html string) {
	el := ComponentRoot(componentID)
	if el.IsNull() || el.IsUndefined() {
		return
	}
	id := el.Value.Call("getAttribute", "data-component-id")
	if !id.Truthy() || id.String() != componentID {
		return
	}
	UpdateDOM(componentID, html)
}

// UpdateDOMIn renders html into an explicit target element (the router
// outlet). The subtree is replaced wholesale: across different component
// trees a positional diff would leave stale nodes behind.
func UpdateDOMIn(target Element, componentID, html string) {
	if target.IsNull() || target.IsUndefined() {
		return
	}
	target.Value.Set("innerHTML", html)

	if TemplateHook != nil {
		TemplateHook(componentID, html)
	}

	ReleaseInputBindings(componentID)

	BindStoreInputsForComponent(componentID, target.Value)
	BindSignalInputs(componentID, target.Value)
	BindASTStoreInputs(componentID, target.Value)
	BindASTSignalInputs(componentID, target.Value)
}

// BindASTStoreInputs binds input elements that have data-bind-store attributes
// (emitted by the AST renderer) to their store variables.
func BindASTStoreInputs(componentID string, element js.Value) {
	inputs := element.Call("querySelectorAll", "[data-bind-store]")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		binding := input.Call("getAttribute", "data-bind-store").String()
		parts := strings.Split(binding, ".")
		if len(parts) != 3 {
			continue
		}
		module, storeName, key := parts[0], parts[1], parts[2]
		store := state.GlobalStoreManager.GetStore(module, storeName)
		if store == nil {
			continue
		}
		if StoreBindingHook != nil && componentID != "" {
			StoreBindingHook(componentID, module, storeName, key)
		}
		storeValue := store.Get(key)
		tag := strings.ToLower(input.Get("tagName").String())
		if tag == "input" {
			inputType := input.Get("type").String()
			if inputType == "checkbox" {
				if b, ok := storeValue.(bool); ok {
					input.Set("checked", b)
				}
				ch, stop := events.Listen("change", input)
				addInputBindingStop(componentID, stop)
				go func(in js.Value, st *state.Store, k string) {
					for range ch {
						st.Set(k, in.Get("checked").Bool())
					}
				}(input, store, key)
				continue
			}
		}
		if storeValue == nil {
			storeValue = ""
		}
		input.Set("value", fmt.Sprintf("%v", storeValue))
		ch, stop := events.Listen("input", input)
		addInputBindingStop(componentID, stop)
		go func(in js.Value, st *state.Store, k string) {
			for range ch {
				st.Set(k, in.Get("value").String())
			}
		}(input, store, key)
	}
}

// BindASTSignalInputs binds input elements that have data-bind-signal attributes
// (emitted by the AST renderer) to their signals.
func BindASTSignalInputs(componentID string, element js.Value) {
	inputs := element.Call("querySelectorAll", "[data-bind-signal]")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		name := input.Call("getAttribute", "data-bind-signal").String()
		sig := getSignal(componentID, name)
		if sig == nil {
			continue
		}
		tag := strings.ToLower(input.Get("tagName").String())
		if tag == "input" {
			inputType := input.Get("type").String()
			if inputType == "checkbox" {
				if s, ok := sig.(interface {
					Read() any
					Set(bool)
				}); ok {
					if b, ok := s.Read().(bool); ok {
						input.Set("checked", b)
					}
					ch, stop := events.Listen("change", input)
					addInputBindingStop(componentID, stop)
					go func(in js.Value, sg interface{ Set(bool) }) {
						for range ch {
							sg.Set(in.Get("checked").Bool())
						}
					}(input, s)
					continue
				}
			}
		}
		if s, ok := sig.(interface {
			Read() any
			Set(string)
		}); ok {
			input.Set("value", fmt.Sprintf("%v", s.Read()))
			ch, stop := events.Listen("input", input)
			addInputBindingStop(componentID, stop)
			go func(in js.Value, sg interface{ Set(string) }) {
				for range ch {
					sg.Set(in.Get("value").String())
				}
			}(input, s)
		}
	}
}

// BindStoreInputsForComponent binds input elements to store variables while
// providing the component context for runtime hooks.
func BindStoreInputsForComponent(componentID string, element js.Value) {
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)

		valueAttr := input.Get("value").String()
		checkedAttr := ""
		if input.Call("hasAttribute", "checked").Bool() {
			checkedAttr = input.Call("getAttribute", "checked").String()
		}

		re := reStoreWrite
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

		if StoreBindingHook != nil && componentID != "" {
			StoreBindingHook(componentID, module, storeName, key)
		}

		storeValue := store.Get(key)

		if usesChecked {
			boolVal, _ := storeValue.(bool)
			input.Set("checked", boolVal)
			ch, stop := events.Listen("change", input)
			addInputBindingStop(componentID, stop)
			go func(in js.Value, st *state.Store, k string) {
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
		ch, stop := events.Listen("input", input)
		addInputBindingStop(componentID, stop)
		go func(in js.Value, st *state.Store, k string) {
			for range ch {
				st.Set(k, in.Get("value").String())
			}
		}(input, store, key)
	}
}

// BindStoreInputs binds input elements to store variables.
func BindStoreInputs(element js.Value) {
	BindStoreInputsForComponent("", element)
}

// BindSignalInputs binds input elements to local component signals.
func BindSignalInputs(componentID string, element js.Value) {
	inputs := element.Call("querySelectorAll", "input, select, textarea")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)

		valueAttr := input.Get("value").String()
		checkedAttr := ""
		if input.Call("hasAttribute", "checked").Bool() {
			checkedAttr = input.Call("getAttribute", "checked").String()
		}

		re := reSignalWrite
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
				ch, stop := events.Listen("change", input)
				addInputBindingStop(componentID, stop)
				go func(in js.Value, sg interface{ Set(bool) }) {
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
			ch, stop := events.Listen("input", input)
			addInputBindingStop(componentID, stop)
			go func(in js.Value, sg interface{ Set(string) }) {
				for range ch {
					sg.Set(in.Get("value").String())
				}
			}(input, s)
		}
	}
}

func patchInnerHTML(element js.Value, html string) {
	activeEl := js.Global().Get("document").Get("activeElement")
	activeID := ""
	activeSelStart := 0
	activeSelEnd := 0
	if activeEl.Truthy() {
		tag := activeEl.Get("nodeName").String()
		if tag == "INPUT" || tag == "TEXTAREA" || tag == "SELECT" {
			activeID = activeEl.Get("id").String()
			if activeID != "" {
				activeSelStart = activeEl.Get("selectionStart").Int()
				activeSelEnd = activeEl.Get("selectionEnd").Int()
			}
		}
	}

	template := CreateElement("template")
	template.Set("innerHTML", html)
	newContent := template.Get("content")

	patched := false
	firstChild := newContent.Get("firstChild")
	if firstChild.Truthy() && firstChild.Get("nodeName").String() == "ROOT" {
		cid := element.Call("getAttribute", "data-component-id")
		newCid := firstChild.Call("getAttribute", "data-component-id")
		if cid.Truthy() && cid.String() == newCid.String() {
			patchAttributes(element, firstChild)
			patchChildren(element, firstChild)
			patched = true
		}
	}

	if !patched {
		patchChildren(element, newContent)
	}

	if activeID != "" {
		restore := js.Global().Get("document").Call("getElementById", activeID)
		if restore.Truthy() {
			restore.Call("focus")
			restore.Call("setSelectionRange", activeSelStart, activeSelEnd)
		}
	}
}

func patchChildren(oldParent, newParent js.Value) {
	// Snapshot the significant children (elements and non-blank text).
	// Whitespace-only text nodes are formatting noise: pairing them
	// positionally shifts the diff whenever a keyed list grows or a
	// conditional toggles, morphing unrelated siblings into each other.
	oldKids := significantChildren(oldParent)
	newKids := significantChildren(newParent)

	keyed := make(map[string]js.Value)
	for _, child := range oldKids {
		if key := getDataKey(child); key != "" {
			keyed[key] = child
		}
	}

	consumed := make([]bool, len(oldKids))
	oi := 0
	// cursor returns the first unconsumed old node: inserts anchor before it.
	cursor := func() js.Value {
		for i := oi; i < len(oldKids); i++ {
			if !consumed[i] {
				return oldKids[i]
			}
		}
		return js.Null()
	}
	insertAtCursor := func(node js.Value) {
		if ref := cursor(); ref.Truthy() {
			oldParent.Call("insertBefore", node, ref)
		} else {
			oldParent.Call("appendChild", node)
		}
	}

	for _, newChild := range newKids {
		if key := getDataKey(newChild); key != "" {
			if oldChild, ok := keyed[key]; ok {
				patchNode(oldChild, newChild)
				if ref := cursor(); !oldChild.Equal(ref) {
					insertAtCursor(oldChild)
				} else {
					// already in position: consume it
					for i := oi; i < len(oldKids); i++ {
						if oldKids[i].Equal(oldChild) {
							consumed[i] = true
							break
						}
					}
				}
				delete(keyed, key)
			} else {
				insertAtCursor(newChild.Call("cloneNode", true))
			}
			continue
		}

		// advance past keyed leftovers (handled through the map above)
		for oi < len(oldKids) && (consumed[oi] || getDataKey(oldKids[oi]) != "") {
			oi++
		}
		if oi < len(oldKids) && samePatchType(oldKids[oi], newChild) {
			patchNode(oldKids[oi], newChild)
			consumed[oi] = true
			oi++
		} else if oi < len(oldKids) {
			oldParent.Call("replaceChild", newChild.Call("cloneNode", true), oldKids[oi])
			consumed[oi] = true
			oi++
		} else {
			oldParent.Call("appendChild", newChild.Call("cloneNode", true))
		}
	}

	// leftover keyed nodes not reused by the new render
	for _, child := range keyed {
		child.Call("remove")
	}
	// leftover unkeyed significant nodes past the new list
	for i := 0; i < len(oldKids); i++ {
		if !consumed[i] && getDataKey(oldKids[i]) == "" {
			oldKids[i].Call("remove")
		}
	}
}

// significantChildren returns the child nodes that participate in diffing:
// elements and text nodes with non-whitespace content.
func significantChildren(parent js.Value) []js.Value {
	children := parent.Get("childNodes")
	out := make([]js.Value, 0, children.Length())
	for i := 0; i < children.Length(); i++ {
		child := children.Index(i)
		if child.Get("nodeType").Int() == 3 && strings.TrimSpace(child.Get("nodeValue").String()) == "" {
			continue
		}
		out = append(out, child)
	}
	return out
}

// samePatchType reports whether two nodes may be patched in place: same node
// name and, for conditional wrappers, the same data-condition identity (a
// wrapper morphing into an unrelated sibling emptied whole sections).
func samePatchType(oldNode, newNode js.Value) bool {
	if oldNode.Get("nodeName").String() != newNode.Get("nodeName").String() {
		return false
	}
	if oldNode.Get("nodeType").Int() != 1 {
		return true
	}
	oc := oldNode.Call("getAttribute", "data-condition")
	nc := newNode.Call("getAttribute", "data-condition")
	os, ns := "", ""
	if !oc.IsNull() {
		os = oc.String()
	}
	if !nc.IsNull() {
		ns = nc.String()
	}
	return os == ns
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

	if nodeType == 1 { // Element node
		patchAttributes(oldNode, newNode)
	}
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
