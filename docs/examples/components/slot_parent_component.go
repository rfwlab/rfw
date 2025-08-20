//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/slot_parent_component.rtml
var slotParentComponentTpl []byte

func NewSlotParentComponent(props map[string]any) *core.HTMLComponent {
	c := core.NewComponent("SlotParentComponent", slotParentComponentTpl, props)
	userCardWithSlots := NewSlotChildComponent(nil)
	c.AddDependency("userCard", userCardWithSlots)

	userCardFallback := NewSlotChildComponent(nil)
	c.AddDependency("userCardFallback", userCardFallback)

	return c
}
