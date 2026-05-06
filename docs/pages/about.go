//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/plugins/seo"
)

// About renders the about page.
func About() core.Component {
	seo.SetTitle("About")
	seo.SetMeta("description", "About rfw")
	return core.NewComponent("AboutPage", []byte("<div>About Page</div>"), nil)
}
