//go:build js && wasm

package plugin

import (
	"encoding/json"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/js"
)

// jsPlugin wraps a JavaScript object and adapts it to the core.Plugin interface.
type jsPlugin struct{ v jst.Value }

func (p jsPlugin) Build(json.RawMessage) error { return nil }

// Install wires JavaScript callbacks to the application hooks.
func (p jsPlugin) Install(a *core.App) {
	if fn := p.v.Get("router"); fn.Type() == jst.TypeFunction {
		a.RegisterRouter(func(path string) { fn.Invoke(path) })
	}
	if fn := p.v.Get("store"); fn.Type() == jst.TypeFunction {
		a.RegisterStore(func(module, store, key string, value any) {
			fn.Invoke(module, store, key, jst.ValueOf(value))
		})
	}
	if fn := p.v.Get("template"); fn.Type() == jst.TypeFunction {
		a.RegisterTemplate(func(id, html string) { fn.Invoke(id, html) })
	}
	mount := p.v.Get("mount")
	unmount := p.v.Get("unmount")
	if mount.Type() == jst.TypeFunction || unmount.Type() == jst.TypeFunction {
		a.RegisterLifecycle(
			func(c core.Component) {
				if mount.Type() == jst.TypeFunction {
					mount.Invoke(c.GetName())
				}
			},
			func(c core.Component) {
				if unmount.Type() == jst.TypeFunction {
					unmount.Invoke(c.GetName())
				}
			},
		)
	}
}

// Register converts a JavaScript plugin definition into a core.Plugin
// and installs it into the application.
func Register(v jst.Value) {
	core.RegisterPlugin(jsPlugin{v: v})
}

func init() {
	// Expose plugin registration for JavaScript callers.
	js.ExposeFunc("rfwRegisterPlugin", func(this jst.Value, args []jst.Value) any {
		if len(args) > 0 {
			Register(args[0])
		}
		return nil
	})
}
