//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
)

//go:embed templates/event_listener_component.rtml
var eventListenerComponentTpl []byte

func NewEventListenerComponent() *core.HTMLComponent {
	c := core.NewComponent("EventListenerComponent", eventListenerComponentTpl, nil)
	// Always start from zero to avoid residual persisted values.
	c.Store.Set("clicks", float64(0))
	c.SetOnMount(func(cmp *core.HTMLComponent) {
		btn := dom.Doc().ByID("clickBtn")
		ch := events.Listen("click", btn.Value)
		go func() {
			for range ch {
				switch v := cmp.Store.Get("clicks").(type) {
				case float64:
					cmp.Store.Set("clicks", v+1)
				case int:
					cmp.Store.Set("clicks", float64(v)+1)
				}
			}
		}()
	})
	return c
}
