//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/slot_parent_component.rtml
var slotParentComponentTpl []byte

type SlotParentComponent struct {
	*core.HTMLComponent
}

func NewSlotParentComponent(props map[string]interface{}) *SlotParentComponent {
	c := &SlotParentComponent{
		HTMLComponent: core.NewHTMLComponent("SlotParentComponent", slotParentComponentTpl, props),
	}
	c.SetComponent(c)
	c.Init(nil)

	header := NewHeaderComponent(map[string]interface{}{"title": "User Card Slots"})
	c.AddDependency("header", header)

	userCardWithSlots := NewSlotChildComponent(nil)
	c.AddDependency("userCard", userCardWithSlots)

	userCardFallback := NewSlotChildComponent(nil)
	c.AddDependency("userCardFallback", userCardFallback)

	return c
}
