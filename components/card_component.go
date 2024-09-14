//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/card_component.html
var cardComponentTemplate []byte

type CardComponent struct {
	*framework.BaseComponent
}

func NewCardComponent(props map[string]interface{}) *CardComponent {
	component := &CardComponent{
		BaseComponent: framework.NewBaseComponent("CardComponent", cardComponentTemplate, props),
	}
	component.Init(nil)
	return component
}
