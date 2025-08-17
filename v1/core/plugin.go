//go:build js && wasm

package core

import (
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

// Plugin defines interface for plugins to register hooks on the App.
type Plugin interface {
	Install(*App)
}

// App maintains registered hooks and exposes helper methods for plugins
// to attach to framework events.
type App struct {
	*hooks
}

type hooks struct {
	routerHooks   []func(string)
	storeHooks    []func(module, store, key string, value interface{})
	templateHooks []func(componentID, html string)
	mountHooks    []func(Component)
	unmountHooks  []func(Component)
}

// newApp creates an App with initialized hook storage.
func newApp() *App {
	return &App{hooks: &hooks{}}
}

// RegisterRouter adds a router navigation hook.
func (a *App) RegisterRouter(fn func(string)) {
	a.routerHooks = append(a.routerHooks, fn)
}

// RegisterStore adds a store mutation hook.
func (a *App) RegisterStore(fn func(module, store, key string, value interface{})) {
	a.storeHooks = append(a.storeHooks, fn)
}

// RegisterTemplate adds a template render hook.
func (a *App) RegisterTemplate(fn func(componentID, html string)) {
	a.templateHooks = append(a.templateHooks, fn)
}

// RegisterLifecycle adds hooks for component mount and unmount.
func (a *App) RegisterLifecycle(mount, unmount func(Component)) {
	if mount != nil {
		a.mountHooks = append(a.mountHooks, mount)
	}
	if unmount != nil {
		a.unmountHooks = append(a.unmountHooks, unmount)
	}
}

var app = newApp()

// RegisterPlugin registers a plugin and allows it to add hooks.
func RegisterPlugin(p Plugin) { p.Install(app) }

// TriggerRouter invokes router hooks with the given path.
func TriggerRouter(path string) {
	for _, h := range app.routerHooks {
		h(path)
	}
}

// TriggerStore invokes store hooks for a mutation.
func TriggerStore(module, store, key string, value interface{}) {
	for _, h := range app.storeHooks {
		h(module, store, key, value)
	}
}

// TriggerTemplate invokes template hooks with rendered HTML for a component.
func TriggerTemplate(componentID, html string) {
	for _, h := range app.templateHooks {
		h(componentID, html)
	}
}

// TriggerMount invokes mount lifecycle hooks.
func TriggerMount(c Component) {
	for _, h := range app.mountHooks {
		h(c)
	}
}

// TriggerUnmount invokes unmount lifecycle hooks.
func TriggerUnmount(c Component) {
	for _, h := range app.unmountHooks {
		h(c)
	}
}

func init() {
	state.StoreHook = TriggerStore
	dom.TemplateHook = TriggerTemplate
}
