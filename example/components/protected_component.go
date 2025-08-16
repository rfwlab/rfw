//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/protected_component.rtml
var protectedComponentTpl []byte

type ProtectedComponent struct {
	*core.HTMLComponent
}

func NewProtectedComponent() *ProtectedComponent {
	c := &ProtectedComponent{
		HTMLComponent: core.NewHTMLComponent("ProtectedComponent", protectedComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Protected Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
