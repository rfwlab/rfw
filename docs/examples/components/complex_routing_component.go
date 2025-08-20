//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/complex_routing_component.rtml
var complexRoutingComponentTpl []byte

func NewComplexRoutingComponent() *core.HTMLComponent {
	c := core.NewComponent("ComplexRoutingComponent", complexRoutingComponentTpl, nil)
	return c
}
