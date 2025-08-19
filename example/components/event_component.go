//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
)

//go:embed templates/event_component.rtml
var eventComponentTpl []byte

// EventComponent demonstrates both basic event handling and event modifiers.
// EventComponent demonstrates both basic event handling and event modifiers.
func NewEventComponent() *core.HTMLComponent {
	c := core.NewComponent("EventComponent", eventComponentTpl, nil)

	if c.Store.Get("count") == nil {
		c.Store.Set("count", 0)
	}

	// register increment so buttons in the template can call it
	dom.RegisterHandlerFunc("increment", func() {
		if val, ok := c.Store.Get("count").(int); ok {
			c.Store.Set("count", val+1)
		}
	})

	headerComponent := NewHeaderComponent(map[string]any{
		"title": "Events",
	})
	c.AddDependency("header", headerComponent)

	return c
}
