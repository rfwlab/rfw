//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/another_component.html
var anotherComponentTemplate []byte

type AnotherComponent struct {
	*framework.BaseComponent
}

func NewAnotherComponent() *AnotherComponent {
	component := &AnotherComponent{
		BaseComponent: framework.NewBaseComponent("AnotherComponent", anotherComponentTemplate),
	}
	component.Init(nil)

	headerComponent := NewHeaderComponent()
	component.RegisterChildComponent("header", headerComponent)

	return component
}
