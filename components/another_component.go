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
		BaseComponent: framework.NewBaseComponent("AnotherComponent", anotherComponentTemplate, nil),
	}
	component.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.RegisterChildComponent("header", headerComponent)

	return component
}
