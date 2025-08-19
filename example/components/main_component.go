//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/main_component.rtml
var mainComponentTpl []byte

func NewMainComponent() *core.HTMLComponent {
	c := core.NewComponent("MainComponent", mainComponentTpl, nil)

	cardComponent := NewCardComponent(map[string]any{
		"title": "just a card",
	})
	c.AddDependency("card", cardComponent)

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Main Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
