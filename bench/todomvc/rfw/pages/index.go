//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"

	"github.com/rfwlab/bench/todomvc/components"
)

// Index renders the TodoMVC page.
func Index() core.Component {
	return components.NewTodoComponent()
}
