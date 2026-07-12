//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"

	"github.com/rfwlab/examples/dashboard/components"
)

// About renders the /about page. The pages plugin derives the route from
// the file name at build time.
func About() core.Component {
	return components.NewAboutComponent()
}
