//go:build js && wasm

package components

import (
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
	dom.RegisterHandlerFunc("increment", c.Increment)
	return c
}

func (c *StateManagementComponent) Init(store *state.Store) {
	c.HTMLComponent.Init(store)
	if store.Get("double") == nil {
		state.Map(store, "double", "count", func(v int) int {
			return v * 2
		})
	}
}

func (c *StateManagementComponent) Increment() {
	if v, ok := c.Store.Get("count").(int); ok {
		c.Store.Set("count", v+1)
	}
}
