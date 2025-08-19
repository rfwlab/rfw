//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/list_component.rtml
var listComponentTpl []byte

type ListComponent struct {
	*core.HTMLComponent
}

func NewListComponent(items []core.Component) *ListComponent {
	props := map[string]interface{}{"items": items}
	c := &ListComponent{}
	c.HTMLComponent = core.NewComponentWith("ListComponent", listComponentTpl, props, c)
	return c
}
