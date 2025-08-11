//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/main_component.rtml
var mainComponentTpl []byte

type MainComponent struct {
	*core.HTMLComponent
}

func NewMainComponent() *MainComponent {
	c := &MainComponent{
		HTMLComponent: core.NewHTMLComponent("MainComponent", mainComponentTpl, nil),
	}
	c.Init(nil)

	cardComponent := NewCardComponent(map[string]interface{}{
		"title": "just a card",
	})
	c.AddDependency("card", cardComponent)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Main Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
