//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/red_cube_component.rtml
var redCubeComponentTpl []byte

func NewRedCubeComponent() *core.HTMLComponent {
	return core.NewComponent("RedCubeComponent", redCubeComponentTpl, nil)
}
