//go:build js && wasm

package plugins

import (
	_ "embed"
	"syscall/js"

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

func (c *PluginsComponent) OnMount() {
	js.Global().Get("document").Call("getElementById", "hello").Set("innerText", js.Global().Get("t").Invoke("hello").String())
}
func (c *PluginsComponent) OnUnmount() {}
