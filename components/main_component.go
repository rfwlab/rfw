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

func NewMainComponent() *MyComponent {
	component := &MyComponent{
		BaseComponent: framework.NewBaseComponent("MainComponent", mainComponentTemplate),
	}
	component.Init()

	headerComponent := NewHeaderComponent()
	component.RegisterChildComponent("header", headerComponent)

	return component
}
