//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/complex_routing_component.rtml
var complexRoutingComponentTpl []byte

type ComplexRoutingComponent struct {
	*core.HTMLComponent
}

func NewComplexRoutingComponent() *ComplexRoutingComponent {
	c := &ComplexRoutingComponent{}
	c.HTMLComponent = core.NewComponentWith("ComplexRoutingComponent", complexRoutingComponentTpl, nil, c)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Complex Routing",
	})
	c.AddDependency("header", headerComponent)

	return c
}
