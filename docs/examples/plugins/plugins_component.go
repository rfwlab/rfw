//go:build js && wasm

package plugins

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/plugins/i18n"
)

//go:embed templates/plugins_component.rtml
var pluginsComponentTpl []byte

func NewPluginsComponent() *core.HTMLComponent {
	hello := i18n.Signal("hello")
	cmp := composition.Wrap(core.NewComponent("PluginsComponent", pluginsComponentTpl, nil))
	cmp.Prop("hello", hello)
	cmp.On("setEN", func() { i18n.SetLang("en") })
	cmp.On("setIT", func() { i18n.SetLang("it") })
	return cmp.Unwrap()
}
