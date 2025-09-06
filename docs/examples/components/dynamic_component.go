//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/dynamic_component.rtml
var dynamicComponentTpl []byte

func init() {
	core.MustRegisterComponent("red-cube", func() core.Component { return NewRedCubeComponent() })
}

func NewDynamicComponent() *core.HTMLComponent {
	c := core.NewComponent("DynamicComponent", dynamicComponentTpl, nil)
	list := NewListComponent([]core.Component{
		NewHeaderComponent(map[string]any{"title": "First"}),
		NewHeaderComponent(map[string]any{"title": "Second"}),
	})
	c.AddDependency("list", list)

	return c
}
