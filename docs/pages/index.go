//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/docs/components"
	"github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/plugins/seo"
)

// Index renders the home page.
func Index() core.Component {
	seo.SetTitle("Docs")
	seo.SetMeta("description", "RFW documentation")
	return components.NewHomeComponent()
}
