//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/another_component.html
var anotherComponentTpl []byte

type AnotherComponent struct {
	*framework.BaseComponent
}

func NewAnotherComponent() *AnotherComponent {
	component := &AnotherComponent{
		BaseComponent: framework.NewBaseComponent("AnotherComponent", anotherComponentTpl, nil),
	}
	component.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}
