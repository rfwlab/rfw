//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/state_bindings_component.rtml
var stateBindingsComponentTpl []byte

type StateBindingsComponent struct {
	*core.HTMLComponent
}

func NewStateBindingsComponent() *StateBindingsComponent {
	c := &StateBindingsComponent{}
	c.HTMLComponent = core.NewComponentWith(
		"StateBindingsComponent",
		stateBindingsComponentTpl,
		nil,
		c,
	)
	return c
}

func (c *StateBindingsComponent) Init(store *state.Store) {
	c.HTMLComponent.Init(store)
	if store.Get("name") == nil {
		store.Set("name", "")
	}
	if store.Get("agree") == nil {
		store.Set("agree", false)
	}
	if store.Get("bio") == nil {
		store.Set("bio", "")
	}
}
