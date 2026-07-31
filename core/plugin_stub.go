//go:build !js || !wasm

package core

import "encoding/json"

// Plugin is a no-op stub for non-WASM builds.
type Plugin interface {
	Build(json.RawMessage) error
	Install(*App)
}

// Named exposes a plugin name.
type Named interface{ Name() string }

// Requires lists mandatory plugin dependencies.
type Requires interface{ Requires() []Plugin }

// Optional lists optional plugin dependencies.
type Optional interface{ Optional() []Plugin }

// PreBuilder runs before a build.
type PreBuilder interface{ PreBuild(json.RawMessage) error }

// PostBuilder runs after a build.
type PostBuilder interface{ PostBuild(json.RawMessage) error }

// Uninstaller removes plugin resources.
type Uninstaller interface{ Uninstall(*App) }

// App is a stub holder for callbacks.
type App struct{}

// RegisterRouter performs no work outside WASM.
func (a *App) RegisterRouter(func(string)) {}

// RegisterStore performs no work outside WASM.
func (a *App) RegisterStore(func(module, store, key string, value any)) {}

// RegisterLifecycle performs no work outside WASM.
func (a *App) RegisterLifecycle(func(Component), func(Component)) {}

// RegisterTemplate performs no work outside WASM.
func (a *App) RegisterTemplate(func(componentID, html string)) {}

// RegisterRTMLVar performs no work outside WASM.
func (a *App) RegisterRTMLVar(string, string, any) {}

// HasPlugin reports false outside WASM.
func (a *App) HasPlugin(string) bool { return false }

// RegisterPlugin performs no work outside WASM.
func RegisterPlugin(Plugin) {}

// TriggerRouter performs no work outside WASM.
func TriggerRouter(string) {}

// TriggerStore performs no work outside WASM.
func TriggerStore(string, string, string, any) {}

// TriggerMount performs no work outside WASM.
func TriggerMount(Component) {}

// TriggerUnmount performs no work outside WASM.
func TriggerUnmount(Component) {}

// TriggerTemplate performs no work outside WASM.
func TriggerTemplate(string, string) {}

// OnNavigate performs no work outside WASM.
func OnNavigate(func(string)) {}

// OnTemplate performs no work outside WASM.
func OnTemplate(func(componentID, html string)) {}

// RegisterPluginVar performs no work outside WASM.
func RegisterPluginVar(string, string, any) {}
