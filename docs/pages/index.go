//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/docs/components"
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/plugins/seo"
)

// Index renders the home page.
func Index() core.Component {
	seo.SetTitle("Docs")
	seo.SetMeta("description", "rfw documentation")
	return components.NewDocsComponent()
}
