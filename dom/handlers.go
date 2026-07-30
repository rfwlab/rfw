//go:build js && wasm

package dom

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	js "github.com/rfwlab/rfw/v2/js"
)

// OnHandlerPanic, if set, is called when a registered handler panics.
// Set this from the core package to wire error overlay recovery.
var OnHandlerPanic func(err any, name string)

var (
	handlerMu                sync.RWMutex
	handlerRegistry          = make(map[string]js.Func)
	componentHandlerRegistry = make(map[string]map[string]js.Func)
	eventBindingSeq          atomic.Uint64
)

// RegisterHandler registers a Go function with custom arguments in the handler registry.
// If a handler with the same name already exists, the old wrapper is released.
func RegisterHandler(name string, fn func(this js.Value, args []js.Value) any) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	if old, ok := handlerRegistry[name]; ok {
		old.Release()
	}
	handlerRegistry[name] = js.SafeFuncOf(fn)
}

// RegisterComponentHandler registers a handler owned by one component instance.
func RegisterComponentHandler(componentID, name string, fn func(this js.Value, args []js.Value) any) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	if componentHandlerRegistry[componentID] == nil {
		componentHandlerRegistry[componentID] = make(map[string]js.Func)
	}
	if old, ok := componentHandlerRegistry[componentID][name]; ok {
		old.Release()
	}
	componentHandlerRegistry[componentID][name] = js.SafeFuncOf(fn)
}

// RegisterHandlerFunc registers a no-argument Go function in the handler registry.
func RegisterHandlerFunc(name string, fn func()) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		fn()
		return nil
	})
}

// RegisterComponentHandlerFunc registers a no-argument component handler.
func RegisterComponentHandlerFunc(componentID, name string, fn func()) {
	RegisterComponentHandler(componentID, name, func(this js.Value, args []js.Value) any {
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

// RegisterHandlerElem registers a handler that receives the element carrying
// the data-on-* attribute (resolved by event delegation, so it works for
// markup injected at runtime) together with the event. This is the idiomatic
// way to handle clicks on list rows: render rows with data-on-click="name"
// and read the row's data-* attributes from el.
func RegisterHandlerElem(name string, fn func(el Element, evt Event)) {
	RegisterHandler(name, func(this js.Value, args []js.Value) any {
		var evt, el js.Value
		if len(args) > 0 {
			evt = args[0]
		}
		if len(args) > 1 {
			el = args[1]
		} else if evt.Truthy() {
			el = evt.Get("target")
		}
		fn(Element{el}, Event{evt})
		return nil
	})
}

// GetHandler retrieves a registered handler by name.
func GetHandler(name string) js.Func {
	handlerMu.RLock()
	defer handlerMu.RUnlock()
	if v, ok := handlerRegistry[name]; ok {
		return v
	}
	return js.Func{}
}

// GetComponentHandler resolves a component handler before the global fallback.
func GetComponentHandler(componentID, name string) js.Func {
	handlerMu.RLock()
	defer handlerMu.RUnlock()
	if handlers := componentHandlerRegistry[componentID]; handlers != nil {
		if v, ok := handlers[name]; ok {
			return v
		}
	}
	return handlerRegistry[name]
}

// ReleaseComponentHandlers releases every handler owned by a component.
func ReleaseComponentHandlers(componentID string) {
	handlerMu.Lock()
	handlers := componentHandlerRegistry[componentID]
	delete(componentHandlerRegistry, componentID)
	handlerMu.Unlock()
	for _, handler := range handlers {
		handler.Release()
	}
}

type delegatedHandler struct {
	event   string
	capture bool
	fn      js.Func
	stop    func()
}

// DelegateEvents attaches delegated event listeners on the component root
// element. Bubbling events bubble up to root where data-on-* attributes
// are resolved to registered handlers.
//
// Delegating twice for the same component (a remount, a root replaced by a
// re-render of the surrounding markup) replaces the previous set: keeping it
// would fire every handler twice and leak one js.Func per event per remount.
func DelegateEvents(componentID string, root js.Value) {
	RemoveDelegatedEvents(componentID, root)

	var handlers []delegatedHandler
	events := []string{"click", "submit", "input", "change", "keydown", "keyup", "focus", "blur"}
	for _, evtName := range events {
		for _, capture := range []bool{false, true} {
			if (evtName == "focus" || evtName == "blur") && !capture {
				continue
			}
			handler := newDelegatedHandler(componentID, root, evtName, capture)
			handlers = append(handlers, handler)
			root.Call("addEventListener", evtName, handler.fn, capture)
		}
	}
	delegateMu.Lock()
	delegates[componentID] = handlers
	delegateMu.Unlock()
}

func newDelegatedHandler(componentID string, root js.Value, event string, capture bool) delegatedHandler {
	var timerMu sync.Mutex
	timers := make(map[string]*time.Timer)
	throttled := make(map[string]time.Time)

	fn := js.SafeFuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		evt := args[0]
		target := evt.Get("target")
		for target.Truthy() {
			handlerName := target.Call("getAttribute", "data-on-"+event)
			if handlerName.Truthy() {
				modifiers := eventModifiers(target, event)
				_, wantsCapture := modifiers["capture"]
				nonBubbling := event == "focus" || event == "blur"
				if (nonBubbling || wantsCapture == capture) && eventAllowed(evt, target, modifiers) {
					h := GetComponentHandler(componentID, handlerName.String())
					if h.Truthy() {
						if _, ok := modifiers["prevent"]; ok {
							if _, passive := modifiers["passive"]; !passive {
								evt.Call("preventDefault")
							}
						}
						if _, ok := modifiers["stop"]; ok {
							evt.Call("stopPropagation")
						}
						if _, ok := modifiers["once"]; ok {
							target.Call("removeAttribute", "data-on-"+event)
							target.Call("removeAttribute", "data-on-"+event+"-modifiers")
						}

						invoke := func() {
							defer func() {
								if r := recover(); r != nil && OnHandlerPanic != nil {
									OnHandlerPanic(r, handlerName.String())
								}
							}()
							h.Invoke(evt, target)
						}
						key := eventBindingKey(target, event, handlerName.String())
						if delay, ok := modifierDelay(modifiers, "debounce"); ok {
							timerMu.Lock()
							if timer := timers[key]; timer != nil {
								timer.Stop()
							}
							var scheduled *time.Timer
							scheduled = time.AfterFunc(delay, func() {
								invoke()
								timerMu.Lock()
								if timers[key] == scheduled {
									delete(timers, key)
								}
								timerMu.Unlock()
							})
							timers[key] = scheduled
							timerMu.Unlock()
							return nil
						}
						if delay, ok := modifierDelay(modifiers, "throttle"); ok {
							timerMu.Lock()
							last := throttled[key]
							if time.Since(last) < delay {
								timerMu.Unlock()
								return nil
							}
							throttled[key] = time.Now()
							timerMu.Unlock()
						}
						invoke()
						return nil
					}
				}
			}
			if target.Equal(root) {
				break
			}
			target = target.Get("parentElement")
		}
		return nil
	})

	stop := func() {
		timerMu.Lock()
		for _, timer := range timers {
			timer.Stop()
		}
		clear(timers)
		clear(throttled)
		timerMu.Unlock()
	}
	return delegatedHandler{event: event, capture: capture, fn: fn, stop: stop}
}

func eventModifiers(target js.Value, event string) map[string]struct{} {
	raw := target.Call("getAttribute", "data-on-"+event+"-modifiers")
	modifiers := make(map[string]struct{})
	if !raw.Truthy() {
		return modifiers
	}
	for _, modifier := range strings.Split(raw.String(), ",") {
		modifier = strings.ToLower(strings.TrimSpace(modifier))
		if modifier != "" {
			modifiers[modifier] = struct{}{}
		}
	}
	return modifiers
}

func eventAllowed(evt, target js.Value, modifiers map[string]struct{}) bool {
	if _, ok := modifiers["self"]; ok && !evt.Get("target").Equal(target) {
		return false
	}
	keys := map[string]string{
		"enter": "Enter", "escape": "Escape", "tab": "Tab", "space": " ",
		"up": "ArrowUp", "down": "ArrowDown", "left": "ArrowLeft", "right": "ArrowRight",
	}
	for modifier, key := range keys {
		if _, ok := modifiers[modifier]; ok && evt.Get("key").String() != key {
			return false
		}
	}
	system := map[string]string{"ctrl": "ctrlKey", "shift": "shiftKey", "alt": "altKey", "meta": "metaKey"}
	for modifier, property := range system {
		if _, ok := modifiers[modifier]; ok && !evt.Get(property).Bool() {
			return false
		}
	}
	if _, exact := modifiers["exact"]; exact {
		for modifier, property := range system {
			_, required := modifiers[modifier]
			if evt.Get(property).Bool() != required {
				return false
			}
		}
	}
	return true
}

func modifierDelay(modifiers map[string]struct{}, name string) (time.Duration, bool) {
	if _, ok := modifiers[name]; !ok {
		return 0, false
	}
	delay := 300
	for modifier := range modifiers {
		if ms, err := strconv.Atoi(modifier); err == nil && ms >= 0 {
			delay = ms
			break
		}
	}
	return time.Duration(delay) * time.Millisecond, true
}

func eventBindingKey(target js.Value, event, handler string) string {
	const property = "__rfwEventBinding"
	id := target.Get(property)
	if !id.Truthy() {
		value := strconv.FormatUint(eventBindingSeq.Add(1), 10)
		target.Set(property, value)
		id = target.Get(property)
	}
	return id.String() + ":" + event + ":" + handler
}

// RemoveDelegatedEvents removes all delegated event listeners for the given component.
func RemoveDelegatedEvents(componentID string, root js.Value) {
	delegateMu.Lock()
	handlers, ok := delegates[componentID]
	if ok {
		delete(delegates, componentID)
	}
	delegateMu.Unlock()
	if !ok {
		return
	}
	// A root that is already gone (its subtree was replaced) cannot have its
	// listeners detached, but the callbacks still have to be released.
	live := root.Truthy()
	for _, handler := range handlers {
		if live {
			root.Call("removeEventListener", handler.event, handler.fn.Value, handler.capture)
		}
		handler.stop()
		handler.fn.Release()
	}
}

var (
	delegateMu sync.Mutex
	delegates  = make(map[string][]delegatedHandler)
)
