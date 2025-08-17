//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	jsa "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/event_component.rtml
var eventComponentTpl []byte

// EventComponent demonstrates both basic event handling and event modifiers.
type EventComponent struct {
	*core.HTMLComponent
}

func NewEventComponent() *EventComponent {
	c := &EventComponent{
		HTMLComponent: core.NewHTMLComponent("EventComponent", eventComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	if c.Store.Get("count") == nil {
		c.Store.Set("count", 0)
	}

	// expose increment so buttons in the template can call it
	jsa.Expose("increment", func() {
		if val, ok := c.Store.Get("count").(int); ok {
			c.Store.Set("count", val+1)
		}
	})

	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Events",
	})
	c.AddDependency("header", headerComponent)

	return c
}
