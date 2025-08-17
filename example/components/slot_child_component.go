//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/slot_child_component.rtml
var slotChildComponentTpl []byte

type SlotChildComponent struct {
	*core.HTMLComponent
}

func NewSlotChildComponent(props map[string]interface{}) *SlotChildComponent {
	c := &SlotChildComponent{
		HTMLComponent: core.NewHTMLComponent("SlotChildComponent", slotChildComponentTpl, props),
	}
	c.SetComponent(c)
	c.Init(nil)
	return c
}
