//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/event_listener_component.rtml
var eventListenerComponentTpl []byte

type EventListenerComponent struct {
	*core.HTMLComponent
}

func NewEventListenerComponent() *EventListenerComponent {
	c := &EventListenerComponent{
		HTMLComponent: core.NewHTMLComponent("EventListenerComponent", eventListenerComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	// Always start from zero to avoid residual persisted values.
	c.Store.Set("clicks", float64(0))

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Event Listener Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}

func (c *EventListenerComponent) Mount() {
	c.HTMLComponent.Mount()
	btn := js.Document().Call("getElementById", "clickBtn")
	ch := events.Listen("click", btn)
	go func() {
		for range ch {
			switch v := c.Store.Get("clicks").(type) {
			case float64:
				c.Store.Set("clicks", v+1)
			case int:
				c.Store.Set("clicks", float64(v)+1)
			}
		}
	}()
}
