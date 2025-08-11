//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/card_component.rtml
var cardComponentTpl []byte

type CardComponent struct {
	*core.HTMLComponent
}

func NewCardComponent(props map[string]interface{}) *CardComponent {
	c := &CardComponent{
		HTMLComponent: core.NewHTMLComponent("CardComponent", cardComponentTpl, props),
	}
	c.Init(nil)
	return c
}
