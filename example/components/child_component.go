//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/child_component.rtml
var childComponentTpl []byte

type ChildComponent struct {
	*core.HTMLComponent
}

func NewChildComponent() *ChildComponent {
	c := &ChildComponent{
		HTMLComponent: core.NewHTMLComponent("ChildComponent", childComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Child Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
