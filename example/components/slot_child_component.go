//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/slot_child_component.rtml
var slotChildComponentTpl []byte

func NewSlotChildComponent(props map[string]any) *core.HTMLComponent {
	return core.NewComponent("SlotChildComponent", slotChildComponentTpl, props)
}
