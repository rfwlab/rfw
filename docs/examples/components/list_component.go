//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/list_component.rtml
var listComponentTpl []byte

func NewListComponent(items []core.Component) *core.HTMLComponent {
	props := map[string]any{"items": items}
	return core.NewComponent("ListComponent", listComponentTpl, props)
}
