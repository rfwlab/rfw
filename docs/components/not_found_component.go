//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/not_found_component.rtml
var notFoundTpl []byte

// NewNotFoundComponent returns a simple 404 component.
func NewNotFoundComponent() *core.HTMLComponent {
	return core.NewComponent("NotFoundComponent", notFoundTpl, nil)
}
