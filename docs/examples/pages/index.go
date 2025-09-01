//go:build js && wasm

package pages

import (
    core "github.com/rfwlab/rfw/v1/core"
)

// Index renders the home page.
func Index() core.Component {
    return core.NewComponent("IndexPage", []byte("<div>Home Page</div>"), nil)
}

