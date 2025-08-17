//go:build js && wasm

package js

import (
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/core"
)

// Global returns the JavaScript global object.
func Global() jst.Value {
	return jst.Global()
}

// Expose registers a no-argument Go function under the given name
// on the JavaScript global object.
func Expose(name string, fn func()) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		fn()
		return nil
	}))
}

// ExposeEvent registers a Go function that receives the first argument
// from the JavaScript call as the event object.
func ExposeEvent(name string, fn func(jst.Value)) {
	Global().Set(name, jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		var evt jst.Value
		if len(args) > 0 {
			evt = args[0]
		}
		fn(evt)
		return nil
	}))
}

// ExposeFunc registers a Go function with custom arguments on the
// JavaScript global object.
func ExposeFunc(name string, fn func(this jst.Value, args []jst.Value) interface{}) {
	Global().Set(name, jst.FuncOf(fn))
}

// Stack returns the current JavaScript stack trace using Error().stack.
func Stack() string {
	return Global().Get("Error").New().Get("stack").String()
}

// jsPlugin wraps a JavaScript object and adapts it to the core.Plugin interface.
type jsPlugin struct{ v jst.Value }

// Install wires JavaScript callbacks to the application hooks.
func (p jsPlugin) Install(a *core.App) {
	if fn := p.v.Get("router"); fn.Type() == jst.TypeFunction {
		a.RegisterRouter(func(path string) { fn.Invoke(path) })
	}
	if fn := p.v.Get("store"); fn.Type() == jst.TypeFunction {
		a.RegisterStore(func(module, store, key string, value interface{}) {
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

// RegisterPlugin converts a JavaScript plugin definition into a core.Plugin
// and installs it into the application.
func RegisterPlugin(v jst.Value) {
	core.RegisterPlugin(jsPlugin{v: v})
}

func init() {
	// Expose plugin registration for JavaScript callers.
	Global().Set("rfwRegisterPlugin", jst.FuncOf(func(this jst.Value, args []jst.Value) interface{} {
		if len(args) > 0 {
			RegisterPlugin(args[0])
		}
		return nil
	}))
}
