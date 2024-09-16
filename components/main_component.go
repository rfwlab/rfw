//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/main_component.rtml
var mainComponentTpl []byte

type MainComponent struct {
	*framework.HTMLComponent
}

func NewMainComponent() *MainComponent {
	component := &MainComponent{
		HTMLComponent: framework.NewHTMLComponent("MainComponent", mainComponentTpl, nil),
	}
	component.Init(nil)

	cardComponent := NewCardComponent(map[string]interface{}{
		"title": "just a card",
	})
	component.AddDependency("card", cardComponent)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Main Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
