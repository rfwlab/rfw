//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"

	"github.com/rfwlab/examples/counter/components"
)

// Index renders the home page.
func Index() core.Component {
	return components.NewCounterComponent()
}
