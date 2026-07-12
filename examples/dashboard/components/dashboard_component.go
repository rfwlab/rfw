//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

//go:embed templates/dashboard_component.rtml
var dashboardTpl []byte

type DashboardComponent struct {
	*core.HTMLComponent
	stop chan struct{}
}

func NewDashboardComponent() *DashboardComponent {
	seedMetrics()

	c := &DashboardComponent{
		HTMLComponent: core.NewHTMLComponent("DashboardComponent", dashboardTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("togglePause", togglePause)

	c.SetOnMount(func(*core.HTMLComponent) {
		c.stop = make(chan struct{})
		startFeed(c.stop)
	})
	c.SetOnUnmount(func(*core.HTMLComponent) {
		close(c.stop)
	})

	c.Init(nil)
	return c
}
