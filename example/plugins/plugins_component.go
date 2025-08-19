//go:build js && wasm

package plugins

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/plugins_component.rtml
var pluginsComponentTpl []byte

type PluginsComponent struct {
	*core.HTMLComponent
}

func NewPluginsComponent() *PluginsComponent {
	c := &PluginsComponent{}
	c.HTMLComponent = core.NewComponentWith("PluginsComponent", pluginsComponentTpl, nil, c)
	return c
}

func (c *PluginsComponent) OnMount() {
	js.Document().Call("getElementById", "hello").Set("innerText", js.Get("t").Invoke("hello").String())
}
func (c *PluginsComponent) OnUnmount() {}
