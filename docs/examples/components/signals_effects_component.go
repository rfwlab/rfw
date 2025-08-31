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
	count *state.Signal[int]
}

func NewSignalsEffectsComponent() *SignalsEffectsComponent {
	c := &SignalsEffectsComponent{}
	c.count = state.NewSignal(0)
	c.HTMLComponent = core.NewComponentWith("SignalsEffectsComponent", signalsEffectsComponentTpl, map[string]any{"count": c.count}, c)
	dom.RegisterHandlerFunc("inc", c.Increment)
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
