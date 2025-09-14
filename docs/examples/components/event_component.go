//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/event_component.rtml
var eventComponentTpl []byte

// EventComponent demonstrates both basic event handling and event modifiers.
// EventComponent demonstrates both basic event handling and event modifiers.
func NewEventComponent() *core.HTMLComponent {
	cmp := composition.Wrap(core.NewComponent("EventComponent", eventComponentTpl, nil))

	store := cmp.HTMLComponent.Store
	if store.Get("count") == nil {
		store.Set("count", 0)
	}

	// register increment so buttons in the template can call it
	cmp.On("increment", func() {
		if val, ok := store.Get("count").(int); ok {
			store.Set("count", val+1)
		}
	})
	return cmp.Unwrap()
}
