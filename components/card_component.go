//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/card_component.rtml
var cardComponentTpl []byte

type CardComponent struct {
	*framework.HTMLComponent
}

func NewCardComponent(props map[string]interface{}) *CardComponent {
	component := &CardComponent{
		HTMLComponent: framework.NewHTMLComponent("CardComponent", cardComponentTpl, props),
	}
	component.Init(nil)
	return component
}
