//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/home_component.rtml
var homeTpl []byte

func NewHomeComponent() *core.HTMLComponent {
	c := core.NewComponent("HomeComponent", homeTpl, nil)
	return c
}
