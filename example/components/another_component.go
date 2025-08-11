//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/another_component.rtml
var anotherComponentTpl []byte

type AnotherComponent struct {
	*core.HTMLComponent
}

func NewAnotherComponent() *AnotherComponent {
	c := &AnotherComponent{
		HTMLComponent: core.NewHTMLComponent("AnotherComponent", anotherComponentTpl, nil),
	}
	c.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Another Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}

func (c *AnotherComponent) SetRouteParams(params map[string]string) {
	c.HTMLComponent.SetRouteParams(params)
}
