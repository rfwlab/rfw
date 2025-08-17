//go:build js && wasm

package plugins

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/plugins_component.rtml
var pluginsComponentTpl []byte

type PluginsComponent struct {
	*core.HTMLComponent
}

func NewPluginsComponent() *PluginsComponent {
	c := &PluginsComponent{
		HTMLComponent: core.NewHTMLComponent("PluginsComponent", pluginsComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)
	return c
}

func (c *PluginsComponent) OnMount()   {}
func (c *PluginsComponent) OnUnmount() {}
