//go:build js && wasm

package plugins

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/plugins_component.rtml
var pluginsComponentTpl []byte

func NewPluginsComponent() *core.HTMLComponent {
	c := core.NewComponent("PluginsComponent", pluginsComponentTpl, nil)
	c.SetOnMount(func(cmp *core.HTMLComponent) {
		js.Document().Call("getElementById", "hello").Set("innerText", js.Get("t").Invoke("hello").String())
	})
	return c
}
