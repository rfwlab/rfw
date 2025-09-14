//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v1/composition"
	core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/event_listener_component.rtml
var eventListenerComponentTpl []byte

func NewEventListenerComponent() *core.HTMLComponent {
	c := core.NewComponent("EventListenerComponent", eventListenerComponentTpl, nil)
	cmp := composition.Wrap(c)

	// Register event handler before mount so the template can bind it.
	cmp.On("increment", func() {
		switch v := c.Store.Get("clicks").(type) {
		case float64:
			c.Store.Set("clicks", v+1)
		case int:
			c.Store.Set("clicks", v+1)
		default:
			c.Store.Set("clicks", 1)
		}
	})

	// Initialize store on mount.
	cmp.SetOnMount(func(*core.HTMLComponent) { c.Store.Set("clicks", 0) })

	return c
}
