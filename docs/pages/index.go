//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/docs/components"
	"github.com/rfwlab/rfw/v1/core"
)

// Index renders the home page.
func Index() core.Component {
	return components.NewHomeComponent()
}
