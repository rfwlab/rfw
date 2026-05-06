//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v2/core"
)

//go:embed templates/card_component.rtml
var cardComponentTpl []byte

func NewCardComponent(props map[string]any) *core.HTMLComponent {
	return core.NewComponent("CardComponent", cardComponentTpl, props)
}
