//go:build js && wasm

package pages

import (
	core "github.com/rfwlab/rfw/v1/core"
)

// About renders the about page.
func About() core.Component {
	return core.NewComponent("AboutPage", []byte("<div>About Page</div>"), nil)
}
