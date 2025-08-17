//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/dynamic_component.rtml
var dynamicComponentTpl []byte

type DynamicComponent struct {
	*core.HTMLComponent
}

func NewDynamicComponent() *DynamicComponent {
	c := &DynamicComponent{
		HTMLComponent: core.NewHTMLComponent("DynamicComponent", dynamicComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	header := NewHeaderComponent(map[string]interface{}{"title": "Dynamic Component"})
	c.AddDependency("header", header)

	return c
}
