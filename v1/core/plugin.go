//go:build js && wasm

package core

import (
	"encoding/json"

	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

// Plugin defines interface for plugins to register hooks on the App. Plugins can
// provide a build step which is executed by the CLI before the application is
// run and may also attach runtime hooks through Install.
type Plugin interface {
	Build(json.RawMessage) error
	Install(*App)
}

// PreBuilder allows plugins to execute logic before the CLI build step.
// Implementing this interface is optional.
type PreBuilder interface {
	PreBuild(json.RawMessage) error
}

// PostBuilder allows plugins to execute logic after the CLI build step.
// Implementing this interface is optional.
type PostBuilder interface {
	PostBuild(json.RawMessage) error
}

// Uninstaller allows plugins to clean up previously registered hooks.
// Implementing this interface is optional.
type Uninstaller interface {
	Uninstall(*App)
}

// App maintains registered hooks and exposes helper methods for plugins
// to attach to framework events.
type App struct {
	*hooks
	pluginVars map[string]map[string]any
}

type hooks struct {
	routerHooks   []func(string)
	storeHooks    []func(module, store, key string, value any)
	templateHooks []func(componentID, html string)
	mountHooks    []func(Component)
	unmountHooks  []func(Component)
}

// newApp creates an App with initialized hook storage.
func newApp() *App {
	return &App{hooks: &hooks{}, pluginVars: make(map[string]map[string]any)}
}

// RegisterRouter adds a router navigation hook.
func (a *App) RegisterRouter(fn func(string)) {
	a.routerHooks = append(a.routerHooks, fn)
}

// RegisterStore adds a store mutation hook.
func (a *App) RegisterStore(fn func(module, store, key string, value any)) {
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

// RegisterRTMLVar registers a value that can be referenced from RTML as
// {plugin:NAME.VAR}.
func (a *App) RegisterRTMLVar(plugin, name string, val any) {
	if a.pluginVars == nil {
		a.pluginVars = make(map[string]map[string]any)
	}
	if _, ok := a.pluginVars[plugin]; !ok {
		a.pluginVars[plugin] = make(map[string]any)
	}
	a.pluginVars[plugin][name] = val
}

// getRTMLVar retrieves a registered plugin variable.
func getRTMLVar(plugin, name string) (any, bool) {
	if app.pluginVars == nil {
		return nil, false
	}
	if vars, ok := app.pluginVars[plugin]; ok {
		v, ok := vars[name]
		return v, ok
	}
	return nil, false
}

// RegisterPluginVar is a convenience wrapper for plugins to expose variables.
func RegisterPluginVar(plugin, name string, val any) {
	app.RegisterRTMLVar(plugin, name, val)
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
func TriggerStore(module, store, key string, value any) {
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
