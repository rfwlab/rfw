//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/parent_component.rtml
var parentComponentTpl []byte

func NewParentComponent() *core.HTMLComponent {
	c := core.NewComponent("ParentComponent", parentComponentTpl, nil)
	return c
}
