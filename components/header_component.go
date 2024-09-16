//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/header_component.html
var headerComponentTpl []byte

type HeaderComponent struct {
	*framework.HTMLComponent
}

func NewHeaderComponent(props map[string]interface{}) *HeaderComponent {
	component := &HeaderComponent{
		HTMLComponent: framework.NewHTMLComponent("HeaderComponent", headerComponentTpl, props),
	}
	component.Init(nil)

	return component
}
