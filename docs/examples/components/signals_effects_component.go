//go:build js && wasm

package components

import (
	_ "embed"
	"fmt"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	state "github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/signals_effects_component.rtml
var signalsEffectsComponentTpl []byte

type SignalsEffectsComponent struct {
	*core.HTMLComponent
	count   *state.Signal[int]
	message *state.Signal[string]
	items   *state.Signal[[]string]
}

func NewSignalsEffectsComponent() *SignalsEffectsComponent {
	c := &SignalsEffectsComponent{}
	c.count = state.NewSignal(0)
	c.message = state.NewSignal("")
	c.items = state.NewSignal([]string{})
	c.HTMLComponent = core.NewComponentWith(
		"SignalsEffectsComponent",
		signalsEffectsComponentTpl,
		map[string]any{
			"count":   c.count,
			"message": c.message,
			"items":   c.items,
		},
		c,
	)
	dom.RegisterHandlerFunc("inc", c.Increment)
	dom.RegisterHandlerFunc("add", c.AddMessage)
	state.Effect(func() func() {
		v := c.count.Get()
		fmt.Println("count is", v)
		return nil
	})
	return c
}

func (c *SignalsEffectsComponent) Increment() {
	c.count.Set(c.count.Get() + 1)
}

func (c *SignalsEffectsComponent) AddMessage() {
	msg := c.message.Get()
	c.items.Set(append(c.items.Get(), msg))
	c.message.Set("")
}
