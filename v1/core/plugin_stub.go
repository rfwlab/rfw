//go:build !js || !wasm

package core

import "encoding/json"

// Plugin is a no-op stub for non-WASM builds.
type Plugin interface {
	Build(json.RawMessage) error
	Install(*App)
}

type Named interface{ Name() string }

type PreBuilder interface{ PreBuild(json.RawMessage) error }
type PostBuilder interface{ PostBuild(json.RawMessage) error }
type Uninstaller interface{ Uninstall(*App) }

// App is a stub holder for callbacks.
type App struct{}

func (a *App) RegisterRouter(fn func(string))                              {}
func (a *App) RegisterStore(fn func(module, store, key string, value any)) {}
func (a *App) RegisterLifecycle(mount, unmount func(Component))            {}
func (a *App) RegisterTemplate(fn func(componentID, html string))          {}
func (a *App) RegisterRTMLVar(plugin, name string, val any)                {}
func (a *App) HasPlugin(name string) bool                                  { return false }

func RegisterPlugin(p Plugin)                           {}
func TriggerRouter(path string)                         {}
func TriggerStore(module, store, key string, value any) {}
func TriggerMount(c Component)                          {}
func TriggerUnmount(c Component)                        {}
func TriggerTemplate(componentID, html string)          {}

func RegisterPluginVar(plugin, name string, val any) {}
