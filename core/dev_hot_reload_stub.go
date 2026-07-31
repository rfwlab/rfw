//go:build !js || !wasm

package core

// HTMLComponent is the non-WASM component placeholder.
type HTMLComponent struct {
	ID    string
	Name  string
	scope *Scope
}

// Render returns no markup outside WASM.
func (c *HTMLComponent) Render() string { return "" }

// Mount performs no work outside WASM.
func (c *HTMLComponent) Mount() {}

// Unmount performs no work outside WASM.
func (c *HTMLComponent) Unmount() {}

// OnMount performs no work outside WASM.
func (c *HTMLComponent) OnMount() {}

// OnUnmount performs no work outside WASM.
func (c *HTMLComponent) OnUnmount() {}

// GetName returns the component name.
func (c *HTMLComponent) GetName() string { return c.Name }

// GetID returns the component ID.
func (c *HTMLComponent) GetID() string { return c.ID }

// SetSlots performs no work outside WASM.
func (c *HTMLComponent) SetSlots(map[string]any) {}

// Scope returns the component lifecycle scope.
func (c *HTMLComponent) Scope() *Scope {
	if c.scope == nil || c.scope.Closed() {
		c.scope = NewScope()
	}
	return c.scope
}
