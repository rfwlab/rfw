//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v2/core"
)

//go:embed templates/another_component.rtml
var anotherComponentTpl []byte

func NewAnotherComponent() *core.HTMLComponent {
	c := core.NewComponent("AnotherComponent", anotherComponentTpl, nil)
	return c
}
