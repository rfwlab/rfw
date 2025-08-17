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

	header := NewHeaderComponent(map[string]interface{}{"title": "Slots Demo"})
	c.AddDependency("header", header)

	childWithSlots := NewSlotChildComponent(nil)
	c.AddDependency("slotChild", childWithSlots)

	childWithFallback := NewSlotChildComponent(nil)
	c.AddDependency("slotChildFallback", childWithFallback)

	return c
}
