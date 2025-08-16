//go:build !js || !wasm

package core

import "github.com/rfwlab/rfw/v1/state"

// Plugin is a no-op stub for non-WASM builds.
type Plugin interface{ Register(*Hooks) }

// Hooks is a stub holder for callbacks.
type Hooks struct{}

func (h *Hooks) RegisterRouter(fn func(string))                                      {}
func (h *Hooks) RegisterStore(fn func(module, store, key string, value interface{})) {}
func (h *Hooks) RegisterLifecycle(mount, unmount func(Component))                    {}
func (h *Hooks) RegisterTemplate(fn func(componentID, html string))                  {}

type Component interface{}

func RegisterPlugin(p Plugin)                                   {}
func TriggerRouter(path string)                                 {}
func TriggerStore(module, store, key string, value interface{}) {}
func TriggerMount(c Component)                                  {}
func TriggerUnmount(c Component)                                {}
func TriggerTemplate(componentID, html string)                  {}

func init() { state.StoreHook = TriggerStore }
