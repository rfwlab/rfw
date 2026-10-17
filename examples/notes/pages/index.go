//go:build js && wasm

package pages

import (
	"github.com/rfwlab/examples/notes/components"
	"github.com/rfwlab/rfw/v2/core"
)

func Index() core.Component { return components.NewNotesComponent() }
