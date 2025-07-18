//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/framework"
)

//go:embed templates/another_component.rtml
var anotherComponentTpl []byte

type AnotherComponent struct {
	*framework.HTMLComponent
}

func NewAnotherComponent() *AnotherComponent {
	component := &AnotherComponent{
		HTMLComponent: framework.NewHTMLComponent("AnotherComponent", anotherComponentTpl, nil),
	}
	component.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	component.AddDependency("header", headerComponent)

	return component
}

func (c *AnotherComponent) SetRouteParams(params map[string]string) {
	c.HTMLComponent.SetRouteParams(params)
}
