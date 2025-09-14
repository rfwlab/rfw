//go:build js && wasm

package composition

import (
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

// Store creates a new state store namespaced to the component's ID.
//
// Since: Unreleased.
func (c *Component) Store(name string, opts ...state.StoreOption) *state.Store {
	if name == "" {
		panic("composition.Store: empty name")
	}
	if s := state.GlobalStoreManager.GetStore(c.ID, name); s != nil {
		c.createdStores[name] = struct{}{}
		return s
	}
	opts = append(opts, state.WithModule(c.ID))
	s := state.NewStore(name, opts...)
	state.GlobalStoreManager.RegisterStore(c.ID, name, s)
	c.createdStores[name] = struct{}{}
	return s
}

// History registers undo and redo handlers for the provided store.
// `undo` and `redo` are the handler names used in the template.
//
// Since: Unreleased.
func (c *Component) History(s *state.Store, undo, redo string) {
	if s == nil {
		panic("composition.History: nil store")
	}
	if undo == "" || redo == "" {
		panic("composition.History: empty handler name")
	}
	dom.RegisterHandlerFunc(undo, s.Undo)
	dom.RegisterHandlerFunc(redo, s.Redo)
}
