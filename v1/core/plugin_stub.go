//go:build !js || !wasm

package core

import "encoding/json"

// Plugin is a no-op stub for non-WASM builds.
type Plugin interface {
	Build(json.RawMessage) error
	Install(*App)
}

// App is a stub holder for callbacks.
type App struct{}

func (a *App) RegisterRouter(fn func(string))                              {}
func (a *App) RegisterStore(fn func(module, store, key string, value any)) {}
func (a *App) RegisterLifecycle(mount, unmount func(Component))            {}
func (a *App) RegisterTemplate(fn func(componentID, html string))          {}

type Component any

func RegisterPlugin(p Plugin)                           {}
func TriggerRouter(path string)                         {}
func TriggerStore(module, store, key string, value any) {}
func TriggerMount(c Component)                          {}
func TriggerUnmount(c Component)                        {}
func TriggerTemplate(componentID, html string)          {}
