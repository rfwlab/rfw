//go:build js && wasm

package logging

import (
	"log"

	"github.com/rfwlab/rfw/v1/core"
)

// Plugin provides basic console logging for various hooks.
type Plugin struct{}

// New creates a new logging plugin.
func New() core.Plugin { return &Plugin{} }

// Register attaches logging hooks.
func (p *Plugin) Register(h *core.Hooks) {
	h.RegisterRouter(func(path string) {
		log.Printf("navigate -> %s", path)
	})
	h.RegisterStore(func(module, store, key string, value interface{}) {
		log.Printf("store %s/%s: %s=%v", module, store, key, value)
	})
	h.RegisterLifecycle(
		func(c core.Component) { log.Printf("mount %s", c.GetName()) },
		func(c core.Component) { log.Printf("unmount %s", c.GetName()) },
	)
}
