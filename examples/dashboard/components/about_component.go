//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
)

//go:embed templates/about_component.rtml
var aboutTpl []byte

type AboutComponent struct {
	*core.HTMLComponent
}

func NewAboutComponent() *AboutComponent {
	c := &AboutComponent{
		HTMLComponent: core.NewHTMLComponent("AboutComponent", aboutTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)
	return c
}
