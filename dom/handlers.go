//go:build js && wasm

package dom

import (
	"strings"
	"sync"

	js "github.com/rfwlab/rfw/v2/js"
)

// OnHandlerPanic, if set, is called when a registered handler panics.
// Set this from the core package to wire error overlay recovery.
var OnHandlerPanic func(err any, name string)

var handlerRegistry = make(map[string]js.Func)

// RegisterHandler registers a Go function with custom arguments in the handler registry.
// If a handler with the same name already exists, the old wrapper is released.
func RegisterHandler(name string, fn func(this js.Value, args []js.Value) any) {
	if old, ok := handlerRegistry[name]; ok {
		old.Release()
	}
	handlerRegistry[name] = js.FuncOf(fn)
}

// RegisterHandlerFunc registers a no-argument Go function in the handler registry.
func RegisterHandlerFunc(name string, fn func()) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		fn()
		return nil
	})
}

// RegisterHandlerEvent registers a Go function that receives the first argument as an event object.
func RegisterHandlerEvent(name string, fn func(js.Value)) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		var evt js.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	})
}

// GetHandler retrieves a registered handler by name.
func GetHandler(name string) js.Func {
	if v, ok := handlerRegistry[name]; ok {
		return v
	}
	return js.Func{}
}

// DelegateEvents attaches delegated event listeners on the component root
// element. Bubbling events bubble up to root where data-on-* attributes
// are resolved to registered handlers.
func DelegateEvents(componentID string, root js.Value) {
	var handlers []js.Func
	events := []string{"click", "submit", "input", "change", "keydown", "keyup", "focus", "blur"}
	for _, evtName := range events {
		captured := evtName
		fn := js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			evt := args[0]
			target := evt.Get("target")
			for target.Truthy() {
				if target.Equal(root) {
					break
				}
				ds := target.Get("dataset")
				key := "on" + strings.ToUpper(captured[:1]) + captured[1:]
				handlerName := ds.Get(key)
				if handlerName.Truthy() {
					h := GetHandler(handlerName.String())
					if h.Truthy() {
						defer func() {
							if r := recover(); r != nil {
								if OnHandlerPanic != nil {
									OnHandlerPanic(r, handlerName.String())
								}
							}
						}()
						h.Invoke(evt)
						evt.Call("stopPropagation")
						return nil
					}
				}
				target = target.Get("parentElement")
			}
			return nil
		})
		handlers = append(handlers, fn)
		root.Call("addEventListener", captured, fn, false)
	}
	delegateMu.Lock()
	delegates[componentID] = handlers
	delegateMu.Unlock()
}

// RemoveDelegatedEvents removes all delegated event listeners for the given component.
func RemoveDelegatedEvents(componentID string, root js.Value) {
	delegateMu.Lock()
	handlers, ok := delegates[componentID]
	if ok {
		delete(delegates, componentID)
	}
	delegateMu.Unlock()
	if !ok || !root.Truthy() {
		return
	}
	events := []string{"click", "submit", "input", "change", "keydown", "keyup", "focus", "blur"}
	for i, evtName := range events {
		if i < len(handlers) {
			root.Call("removeEventListener", evtName, handlers[i].Value)
			handlers[i].Release()
		}
	}
}

var (
	delegateMu sync.Mutex
	delegates  = make(map[string][]js.Func)
)
