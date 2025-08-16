//go:build js && wasm

package core

import (
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

// Plugin defines interface for plugins to register hooks.
type Plugin interface {
	Register(*Hooks)
}

// Hooks stores callbacks for router, store, and lifecycle events.
type Hooks struct {
	routerHooks   []func(string)
	storeHooks    []func(module, store, key string, value interface{})
	templateHooks []func(componentID, html string)
	mountHooks    []func(Component)
	unmountHooks  []func(Component)
}

// RegisterRouter adds a router navigation hook.
func (h *Hooks) RegisterRouter(fn func(string)) {
	h.routerHooks = append(h.routerHooks, fn)
}

// RegisterStore adds a store mutation hook.
func (h *Hooks) RegisterStore(fn func(module, store, key string, value interface{})) {
	h.storeHooks = append(h.storeHooks, fn)
}

// RegisterTemplate adds a template render hook.
func (h *Hooks) RegisterTemplate(fn func(componentID, html string)) {
	h.templateHooks = append(h.templateHooks, fn)
}

// RegisterLifecycle adds hooks for component mount and unmount.
func (h *Hooks) RegisterLifecycle(mount, unmount func(Component)) {
	if mount != nil {
		h.mountHooks = append(h.mountHooks, mount)
	}
	if unmount != nil {
		h.unmountHooks = append(h.unmountHooks, unmount)
	}
}

var globalHooks = &Hooks{}

// RegisterPlugin registers a plugin and allows it to add hooks.
func RegisterPlugin(p Plugin) { p.Register(globalHooks) }

// TriggerRouter invokes router hooks with the given path.
func TriggerRouter(path string) {
	for _, h := range globalHooks.routerHooks {
		h(path)
	}
}

// TriggerStore invokes store hooks for a mutation.
func TriggerStore(module, store, key string, value interface{}) {
	for _, h := range globalHooks.storeHooks {
		h(module, store, key, value)
	}
}

// TriggerTemplate invokes template hooks with rendered HTML for a component.
func TriggerTemplate(componentID, html string) {
	for _, h := range globalHooks.templateHooks {
		h(componentID, html)
	}
}

// TriggerMount invokes mount lifecycle hooks.
func TriggerMount(c Component) {
	for _, h := range globalHooks.mountHooks {
		h(c)
	}
}

// TriggerUnmount invokes unmount lifecycle hooks.
func TriggerUnmount(c Component) {
	for _, h := range globalHooks.unmountHooks {
		h(c)
	}
}

func init() {
	state.StoreHook = TriggerStore
	dom.TemplateHook = TriggerTemplate
}
