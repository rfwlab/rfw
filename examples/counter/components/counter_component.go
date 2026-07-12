//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/state"
)

//go:embed templates/counter_component.rtml
var counterTpl []byte

// counter is registered globally as module "app", store "counter",
// so the template references it as @store:app.counter.count.
var counter = state.NewStore("counter", state.WithModule("app"))

type CounterComponent struct {
	*core.HTMLComponent
}

func NewCounterComponent() *CounterComponent {
	counter.Set("count", 0)

	c := &CounterComponent{
		HTMLComponent: core.NewHTMLComponent("CounterComponent", counterTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("increment", func() {
		n, _ := counter.Get("count").(int)
		counter.Set("count", n+1)
	})

	c.Init(nil)
	return c
}
