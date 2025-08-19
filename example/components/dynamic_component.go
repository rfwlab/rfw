//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/dynamic_component.rtml
var dynamicComponentTpl []byte

func NewDynamicComponent() *core.HTMLComponent {
	c := core.NewComponent("DynamicComponent", dynamicComponentTpl, nil)

	header := NewHeaderComponent(map[string]interface{}{"title": "Dynamic Component"})
	c.AddDependency("header", header)

	return c
}
