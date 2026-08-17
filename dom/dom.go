//go:build js && wasm

// Package dom provides utilities for updating the browser DOM and binding
// event listeners for WebAssembly components.
package dom

import (
	"fmt"
	"log"
	"regexp"
	"runtime/debug"
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
	reStoreWrite  = regexp.MustCompile(`@store:([\w-]+)\.([\w-]+)\.([\w-]+):w`)
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

// StoreBindingHook is invoked for each @store binding, when set, associated with
// a component. It receives the component identifier along with the store module,
// store name, and key that are bound in the DOM.
var StoreBindingHook func(componentID, module, store, key string)

type recoveredDOMPanic struct {
	value any
	stack []byte
}

func (p recoveredDOMPanic) Error() string { return fmt.Sprint(p.value) }
func (p recoveredDOMPanic) Stack() []byte { return p.stack }

func recoverDOMUpdate(componentID string) {
	if recovered := recover(); recovered != nil {
		panicValue := recoveredDOMPanic{value: recovered, stack: debug.Stack()}
		if OnHandlerPanic == nil {
			log.Printf("[rfw] recovered DOM update panic for %s: %v\n%s", componentID, recovered, panicValue.stack)
			return
		}
		func() {
			defer func() {
				if hookPanic := recover(); hookPanic != nil {
					log.Printf("[rfw] DOM panic reporter failed: %v", hookPanic)
					log.Printf("[rfw] recovered DOM update panic for %s: %v\n%s", componentID, recovered, panicValue.stack)
				}
			}()
			OnHandlerPanic(panicValue, "DOM update: "+componentID)
		}()
	}
}

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
	defer recoverDOMUpdate(componentID)
	element := ComponentRoot(componentID)
	if element.IsNull() || element.IsUndefined() {
		return
	}
	activeForm := captureActiveFormState(element.Value)

	// Diff-patch only when the resolved element is the component's OWN root: that
	// is an in-place reactive update, where patching preserves focus/selection.
	// Otherwise the target is the #app fallback (a fresh mount or a route change,
	// since ComponentRoot falls back to #app when the component root is not yet
	// in the DOM). There, positionally diffing two different <root> trees leaves
	// stale nodes from the previous component, so replace wholesale instead.
	elID := element.Call("getAttribute", "data-component-id")
	if componentID != "" && elID.Truthy() && elID.String() == componentID {
		patchInnerHTML(element.Value, html)
	} else {
		element.Set("innerHTML", html)
		recordRenderedTree(element.Value)
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
	activeForm.restore()
	UpdateLifecycleHooks(componentID)
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
	id := el.Call("getAttribute", "data-component-id")
	if !id.Truthy() || id.String() != componentID {
		return
	}
	UpdateDOM(componentID, html)
}

// UpdateDOMIn renders html into an explicit target element (the router
// outlet). The subtree is replaced wholesale: across different component
// trees a positional diff would leave stale nodes behind.
func UpdateDOMIn(target Element, componentID, html string) {
	defer recoverDOMUpdate(componentID)
	if target.IsNull() || target.IsUndefined() {
		return
	}
	target.Set("innerHTML", html)
	recordRenderedTree(target.Value)

	if TemplateHook != nil {
		TemplateHook(componentID, html)
	}

	ReleaseInputBindings(componentID)

	BindStoreInputsForComponent(componentID, target.Value)
	BindSignalInputs(componentID, target.Value)
	BindASTStoreInputs(componentID, target.Value)
	BindASTSignalInputs(componentID, target.Value)
	UpdateLifecycleHooks(componentID)
}

// BindASTStoreInputs binds input elements that have data-bind-store attributes
// (emitted by the AST renderer) to their store variables.
func BindASTStoreInputs(componentID string, element js.Value) {
	inputs := element.Call("querySelectorAll", "[data-bind-store]")
	for i := 0; i < inputs.Length(); i++ {
		input := inputs.Index(i)
		if !componentOwnsElement(componentID, input) {
			continue
		}
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
					setCheckedIfChanged(input, b)
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
		setValueIfChanged(input, fmt.Sprintf("%v", storeValue))
		ch, stop := events.Listen("input", input)
		addInputBindingStop(componentID, stop)
		go func(in js.Value, st *state.Store, k string) {
			for event := range ch {
				if inputEventIsComposing(event) {
					continue
				}
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
		if !componentOwnsElement(componentID, input) {
			continue
		}
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
						setCheckedIfChanged(input, b)
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
			setValueIfChanged(input, fmt.Sprintf("%v", s.Read()))
			ch, stop := events.Listen("input", input)
			addInputBindingStop(componentID, stop)
			go func(in js.Value, sg interface{ Set(string) }) {
				for event := range ch {
					if inputEventIsComposing(event) {
						continue
					}
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
		if !componentOwnsElement(componentID, input) {
			continue
		}

		valueAttr := ""
		if input.Call("hasAttribute", "value").Bool() {
			valueAttr = input.Call("getAttribute", "value").String()
		}
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
			setCheckedIfChanged(input, boolVal)
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
		setValueIfChanged(input, fmt.Sprintf("%v", storeValue))
		ch, stop := events.Listen("input", input)
		addInputBindingStop(componentID, stop)
		go func(in js.Value, st *state.Store, k string) {
			for event := range ch {
				if inputEventIsComposing(event) {
					continue
				}
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
		if !componentOwnsElement(componentID, input) {
			continue
		}

		valueAttr := ""
		if input.Call("hasAttribute", "value").Bool() {
			valueAttr = input.Call("getAttribute", "value").String()
		}
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
					setCheckedIfChanged(input, b)
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
			setValueIfChanged(input, fmt.Sprintf("%v", s.Read()))
			ch, stop := events.Listen("input", input)
			addInputBindingStop(componentID, stop)
			go func(in js.Value, sg interface{ Set(string) }) {
				for event := range ch {
					if inputEventIsComposing(event) {
						continue
					}
					sg.Set(in.Get("value").String())
				}
			}(input, s)
		}
	}
}

func componentOwnsElement(componentID string, element js.Value) bool {
	if componentID == "" {
		return true
	}
	root := element.Call("closest", "[data-component-id]")
	return root.Truthy() && attribute(root, "data-component-id") == componentID
}

func setValueIfChanged(element js.Value, value string) {
	if element.Get("value").String() != value {
		element.Set("value", value)
	}
}

func setCheckedIfChanged(element js.Value, checked bool) {
	if element.Get("checked").Bool() != checked {
		element.Set("checked", checked)
	}
}

func inputEventIsComposing(event js.Value) bool {
	value := event.Get("isComposing")
	return value.Type() == js.TypeBoolean && value.Bool()
}
