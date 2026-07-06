//go:build !js || !wasm

package core

type HTMLComponent struct {
	ID   string
	Name string
}

func (c *HTMLComponent) Render() string          { return "" }
func (c *HTMLComponent) Mount()                   {}
func (c *HTMLComponent) Unmount()                 {}
func (c *HTMLComponent) OnMount()                 {}
func (c *HTMLComponent) OnUnmount()               {}
func (c *HTMLComponent) GetName() string          { return c.Name }
func (c *HTMLComponent) GetID() string            { return c.ID }
func (c *HTMLComponent) SetSlots(map[string]any)  {}