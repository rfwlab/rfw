//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/ssc_component.rtml
var sscTpl []byte

func NewSSCComponent() *core.HTMLComponent {
	c := core.NewComponent("SSCComponent", sscTpl, nil)
	c.AddHostComponent("SSCHost")
	return c
}
