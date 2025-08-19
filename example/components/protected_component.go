//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/protected_component.rtml
var protectedComponentTpl []byte

func NewProtectedComponent() *core.HTMLComponent {
	c := core.NewComponent("ProtectedComponent", protectedComponentTpl, nil)

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Protected Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}
