//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/mirkobrombin/rfw/framework"
)

//go:embed templates/header_component.html
var headerComponentTemplate []byte

type HeaderComponent struct {
	*framework.BaseComponent
}

func NewHeaderComponent() *HeaderComponent {
	component := &HeaderComponent{
		BaseComponent: framework.NewBaseComponent("HeaderComponent", headerComponentTemplate),
	}
	component.Init(nil)

	return component
}
