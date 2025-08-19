//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/child_component.rtml
var childComponentTpl []byte

func NewChildComponent() *core.HTMLComponent {
	c := core.NewComponent("ChildComponent", childComponentTpl, nil)

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Child Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
