//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/parent_component.rtml
var parentComponentTpl []byte

type ParentComponent struct {
	*core.HTMLComponent
}

func NewParentComponent() *ParentComponent {
	c := &ParentComponent{
		HTMLComponent: core.NewHTMLComponent("ParentComponent", parentComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Parent Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
