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

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "State Management",
	})
	c.AddDependency("header", headerComponent)

	dom.RegisterHandlerFunc("increment", c.Increment)

	return c
}

func (c *StateManagementComponent) Init(store *state.Store) {
	c.HTMLComponent.Init(store)
	if store.Get("double") == nil {
		store.RegisterComputed(state.NewComputed("double", []string{"count"}, func(m map[string]interface{}) interface{} {
			if v, ok := m["count"].(int); ok {
				return v * 2
			}
			return 0
		}))
	}
}

func (c *StateManagementComponent) Increment() {
	if v, ok := c.Store.Get("count").(int); ok {
		c.Store.Set("count", v+1)
	}
}
