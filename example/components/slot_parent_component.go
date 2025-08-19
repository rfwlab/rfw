//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/slot_parent_component.rtml
var slotParentComponentTpl []byte

func NewSlotParentComponent(props map[string]interface{}) *core.HTMLComponent {
	c := core.NewComponent("SlotParentComponent", slotParentComponentTpl, props)

	header := NewHeaderComponent(map[string]interface{}{"title": "User Card Slots"})
	c.AddDependency("header", header)

	userCardWithSlots := NewSlotChildComponent(nil)
	c.AddDependency("userCard", userCardWithSlots)

	userCardFallback := NewSlotChildComponent(nil)
	c.AddDependency("userCardFallback", userCardFallback)

	return c
}
