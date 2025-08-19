//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/header_component.rtml
var headerComponentTpl []byte

type HeaderComponent struct {
	*core.HTMLComponent
}

func NewHeaderComponent(props map[string]interface{}) *HeaderComponent {
	c := &HeaderComponent{}
	c.HTMLComponent = core.NewComponentWith("HeaderComponent", headerComponentTpl, props, c)

	return c
}

func (c *HeaderComponent) OnMount() {
	if count, ok := c.Store.Get("headerMounts").(int); ok {
		c.Store.Set("headerMounts", count+1)
	} else {
		c.Store.Set("headerMounts", 1)
	}

	if _, ok := c.Store.Get("headerUnmounts").(int); !ok {
		c.Store.Set("headerUnmounts", 0)
	}
}

func (c *HeaderComponent) OnUnmount() {
	if count, ok := c.Store.Get("headerUnmounts").(int); ok {
		c.Store.Set("headerUnmounts", count+1)
	} else {
		c.Store.Set("headerUnmounts", 1)
	}
}
