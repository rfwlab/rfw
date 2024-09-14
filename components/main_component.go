//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/main_component.html
var mainComponentTemplate []byte

type MainComponent struct {
	*framework.BaseComponent
}

func NewMainComponent(title string) *MainComponent {
	component := &MainComponent{
		BaseComponent: framework.NewBaseComponent("MainComponent", mainComponentTemplate, nil),
	}
	component.Init(nil)

	cardComponent := NewCardComponent(map[string]interface{}{
		"title": title,
	})
	component.RegisterChildComponent("card", cardComponent)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": title,
	})
	component.RegisterChildComponent("header", headerComponent)

	return component
}
