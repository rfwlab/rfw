//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/header_component.rtml
var headerComponentTpl []byte

type HeaderComponent struct {
	*core.HTMLComponent
}

func NewHeaderComponent(props map[string]interface{}) *HeaderComponent {
	c := &HeaderComponent{
		HTMLComponent: core.NewHTMLComponent("HeaderComponent", headerComponentTpl, props),
	}
	c.Init(nil)

	return c
}
