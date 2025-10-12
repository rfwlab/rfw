//go:build js && wasm && rfwdev

package core

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/rfwlab/rfw/v1/dom"
	js "github.com/rfwlab/rfw/v1/js"
)

type devPayload struct {
	Type      string `json:"type"`
	Component string `json:"component,omitempty"`
	Markup    string `json:"markup,omitempty"`
}

var (
	devWatcherOnce sync.Once
	devEventSource js.Value
	devMsgHandler  js.Func
	devErrHandler  js.Func

	devMu                sync.RWMutex
	devComponents        = map[string]map[string]*HTMLComponent{}
	devTemplateOverrides = map[string]string{}
)

func startDevTemplateWatcher() {
	devWatcherOnce.Do(func() {
		if js.Get("EventSource").Type() != js.TypeFunction {
			log.Printf("rfw: EventSource not available, auto reload disabled")
			return
		}
		devMsgHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			data := args[0].Get("data").String()
			devHandleDevMessage(data)
			return nil
		})
		devErrHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
			params := make([]any, 0, len(args)+1)
			params = append(params, "rfw dev watcher error")
			for _, a := range args {
				params = append(params, a)
			}
			js.Console().Call("warn", params...)
			return nil
		})
		devEventSource = js.Get("EventSource").New("/__rfw/hmr")
		if !devEventSource.Truthy() {
			return
		}
		if devMsgHandler.Truthy() {
			devEventSource.Set("onmessage", devMsgHandler.Value)
			devEventSource.Call("addEventListener", "message", devMsgHandler.Value)
		}
		if devErrHandler.Truthy() {
			devEventSource.Set("onerror", devErrHandler.Value)
			devEventSource.Call("addEventListener", "error", devErrHandler.Value)
		}
	})
}

func devHandleDevMessage(raw string) {
	var msg devPayload
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		log.Printf("rfw: failed to parse dev message: %v", err)
		return
	}
	switch msg.Type {
	case "reload":
		js.Location().Call("reload")
	case "rtml":
		devApplyTemplateUpdate(msg.Component, msg.Markup)
	}
}

func devApplyTemplateUpdate(name, markup string) {
	if !DevMode || name == "" || markup == "" {
		return
	}
	devMu.Lock()
	devTemplateOverrides[name] = markup
	components := make([]*HTMLComponent, 0, len(devComponents[name]))
	for _, cmp := range devComponents[name] {
		components = append(components, cmp)
	}
	devMu.Unlock()

	dom.OverrideBindings(name, markup)
	for _, cmp := range components {
		cmp.Template = markup
		cmp.TemplateFS = []byte(markup)
		cmp.cache = nil
		cmp.lastCacheKey = ""
		dom.RegisterBindings(cmp.ID, cmp.Name, markup)
		html := cmp.Render()
		dom.UpdateDOM(cmp.ID, html)
	}
}

func devOverrideTemplate(c *HTMLComponent, template string) string {
	if !DevMode {
		return template
	}
	devMu.RLock()
	override, ok := devTemplateOverrides[c.Name]
	devMu.RUnlock()
	if !ok || override == "" {
		return template
	}
	dom.OverrideBindings(c.Name, override)
	c.TemplateFS = []byte(override)
	return override
}

func devRegisterComponent(c *HTMLComponent) {
	if !DevMode {
		return
	}
	devMu.Lock()
	defer devMu.Unlock()
	bucket, ok := devComponents[c.Name]
	if !ok {
		bucket = make(map[string]*HTMLComponent)
		devComponents[c.Name] = bucket
	}
	bucket[c.ID] = c
}

func devUnregisterComponent(c *HTMLComponent) {
	if !DevMode {
		return
	}
	devMu.Lock()
	defer devMu.Unlock()
	if bucket, ok := devComponents[c.Name]; ok {
		delete(bucket, c.ID)
		if len(bucket) == 0 {
			delete(devComponents, c.Name)
		}
	}
}
