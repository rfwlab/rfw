//go:build js && wasm

package soccer

import (
	"encoding/json"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
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
		doc := dom.Doc()
		badges := doc.QueryAll("[data-plugin=\"soccer.badge\"]")
		for i := 0; i < badges.Length(); i++ {
			badges.Index(i).SetText("⚽ Lions FC")
		}
		cmds := doc.QueryAll("[data-plugin-cmd=\"soccer.log\"]")
		for i := 0; i < cmds.Length(); i++ {
			cmds.Index(i).OnClick(func(dom.Event) {
				core.Log().Info("Go Lions!")
			})
		}
	})
}
