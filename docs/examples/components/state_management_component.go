//go:build js && wasm

package components

import (
	"context"
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/state_management_component.rtml
var stateManagementComponentTpl []byte

type StateManagementComponent struct {
	*core.HTMLComponent
}

func NewStateManagementComponent() *StateManagementComponent {
	c := &StateManagementComponent{}
	c.HTMLComponent = core.NewComponentWith("StateManagementComponent", stateManagementComponentTpl, nil, c)
	return c
}

func (c *StateManagementComponent) Init(store *state.Store) {
	c.HTMLComponent.Init(store)
	if store.Get("double") == nil {
		state.Map(store, "double", "count", func(v int) int {
			return v * 2
		})
	}

	increment := state.Action(func(ctx state.Context) error {
		if v, ok := store.Get("count").(int); ok {
			store.Set("count", v+1)
		}
		return nil
	})
	handler := state.UseAction(context.Background(), increment)
	dom.RegisterHandlerFunc("increment", func() { _ = handler() })
}
