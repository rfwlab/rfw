//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	dom "github.com/rfwlab/rfw/v1/dom"
	events "github.com/rfwlab/rfw/v1/events"
	js "github.com/rfwlab/rfw/v1/js"
)

//go:embed templates/observer_component.rtml
var observerComponentTpl []byte

type ObserverComponent struct {
	*core.HTMLComponent
	stopMut func()
	stopInt func()
}

func NewObserverComponent() *ObserverComponent {
	c := &ObserverComponent{}
	c.HTMLComponent = core.NewComponentWith("ObserverComponent", observerComponentTpl, nil, c)

	// Reset counts to zero to provide a predictable demo state.
	c.Store.Set("mutations", float64(0))
	c.Store.Set("intersections", float64(0))
	return c
}

func (c *ObserverComponent) Mount() {
	c.HTMLComponent.Mount()
	doc := dom.Doc()
	mutCh, stopMut := events.ObserveMutations("#observeMe")
	c.stopMut = stopMut
	go func() {
		for range mutCh {
			switch v := c.Store.Get("mutations").(type) {
			case float64:
				c.Store.Set("mutations", v+1)
			case int:
				c.Store.Set("mutations", float64(v)+1)
			}
		}
	}()

	// Button to mutate the observed node and trigger MutationObserver.
	mutateBtn := doc.ByID("mutateBtn")
	mutateCh := events.Listen("click", mutateBtn.Value)
	go func() {
		for range mutateCh {
			node := doc.ByID("observeMe")
			child := doc.CreateElement("span")
			node.Call("appendChild", child.Value)
		}
	}()

	opts := js.NewDict()
	intCh, stopInt := events.ObserveIntersections(".watched", opts.Value)
	c.stopInt = stopInt
	go func() {
		for range intCh {
			switch v := c.Store.Get("intersections").(type) {
			case float64:
				c.Store.Set("intersections", v+1)
			case int:
				c.Store.Set("intersections", float64(v)+1)
			}
		}
	}()

	// Button to toggle visibility and trigger IntersectionObserver.
	toggleBtn := doc.ByID("toggleBtn")
	toggleCh := events.Listen("click", toggleBtn.Value)
	go func() {
		for range toggleCh {
			el := doc.Query(".watched")
			style := el.Get("style")
			if style.Get("display").String() == "none" {
				style.Set("display", "block")
			} else {
				style.Set("display", "none")
			}
		}
	}()
}

func (c *ObserverComponent) Unmount() {
	if c.stopMut != nil {
		c.stopMut()
	}
	if c.stopInt != nil {
		c.stopInt()
	}
	c.HTMLComponent.Unmount()
}
