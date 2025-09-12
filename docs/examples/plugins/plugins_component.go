//go:build js && wasm

package plugins

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/plugins/i18n"
)

//go:embed templates/plugins_component.rtml
var pluginsComponentTpl []byte

type PluginsComponent struct{ *core.HTMLComponent }

func NewPluginsComponent() *PluginsComponent {
	hello := i18n.Signal("hello")
	c := &PluginsComponent{}
	c.HTMLComponent = core.NewComponentWith(
		"PluginsComponent",
		pluginsComponentTpl,
		map[string]any{"hello": hello},
		c,
	)
	dom.RegisterHandlerFunc("setEN", c.SetEN)
	dom.RegisterHandlerFunc("setIT", c.SetIT)
	return c
}

func (c *PluginsComponent) SetEN() { i18n.SetLang("en") }
func (c *PluginsComponent) SetIT() { i18n.SetLang("it") }
