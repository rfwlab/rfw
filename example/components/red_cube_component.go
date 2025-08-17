//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/red_cube_component.rtml
var redCubeComponentTpl []byte

type RedCubeComponent struct {
	*core.HTMLComponent
}

func NewRedCubeComponent() *RedCubeComponent {
	c := &RedCubeComponent{
		HTMLComponent: core.NewHTMLComponent("RedCubeComponent", redCubeComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)
	return c
}
