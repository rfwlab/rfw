//go:build js && wasm

package soccer

import (
	"encoding/json"

	core "github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
)

// Plugin exposes sample soccer data through plugin directives.
type Plugin struct{}

// New creates a new soccer plugin instance.
func New() core.Plugin { return Plugin{} }

// Build is a no-op for the example plugin.
func (Plugin) Build(json.RawMessage) error { return nil }

// Install registers a variable and handles plugin commands/constructors.
func (Plugin) Install(a *core.App) {
	core.RegisterPluginVar("soccer", "team", "Lions")
	a.RegisterTemplate(func(componentID, html string) {
		doc := js.Document()
		badges := doc.Call("querySelectorAll", "[data-plugin=\"soccer.badge\"]")
		for i := 0; i < badges.Get("length").Int(); i++ {
			el := badges.Index(i)
			el.Set("textContent", "⚽ Lions FC")
		}
		cmds := doc.Call("querySelectorAll", "[data-plugin-cmd=\"soccer.log\"]")
		for i := 0; i < cmds.Get("length").Int(); i++ {
			el := cmds.Index(i)
			handler := js.FuncOf(func(this js.Value, args []js.Value) any {
				js.Console().Call("log", "Go Lions!")
				return nil
			})
			el.Call("addEventListener", "click", handler)
		}
	})
}
