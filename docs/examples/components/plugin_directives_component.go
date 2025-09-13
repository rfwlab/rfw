//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/plugin_directives_component.rtml
var pluginDirectivesTpl []byte

type PluginDirectivesComponent struct{ *core.HTMLComponent }

func NewPluginDirectivesComponent() *PluginDirectivesComponent {
	c := &PluginDirectivesComponent{}
	c.HTMLComponent = core.NewComponentWith("PluginDirectivesComponent", pluginDirectivesTpl, nil, c)
	return c
}
