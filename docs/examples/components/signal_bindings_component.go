//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/signal_bindings_component.rtml
var signalBindingsComponentTpl []byte

type SignalBindingsComponent struct {
	*core.HTMLComponent
	name  *state.Signal[string]
	agree *state.Signal[bool]
	bio   *state.Signal[string]
}

func NewSignalBindingsComponent() *SignalBindingsComponent {
	c := &SignalBindingsComponent{}
	c.name = state.NewSignal("")
	c.agree = state.NewSignal(false)
	c.bio = state.NewSignal("")
	c.HTMLComponent = core.NewComponentWith(
		"SignalBindingsComponent",
		signalBindingsComponentTpl,
		map[string]any{
			"name":  c.name,
			"agree": c.agree,
			"bio":   c.bio,
		},
		c,
	)
	return c
}
