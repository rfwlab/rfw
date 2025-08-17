//go:build js && wasm

package components

import (
	_ "embed"

	core "github.com/rfwlab/rfw/v1/core"
	events "github.com/rfwlab/rfw/v1/events"
	jsa "github.com/rfwlab/rfw/v1/js"
	jst "syscall/js"
)

//go:embed templates/observer_component.rtml
var observerComponentTpl []byte

type ObserverComponent struct {
	*core.HTMLComponent
	stopMut func()
	stopInt func()
}

func NewObserverComponent() *ObserverComponent {
	c := &ObserverComponent{
		HTMLComponent: core.NewHTMLComponent("ObserverComponent", observerComponentTpl, nil),
	}
	c.SetComponent(c)
	c.Init(nil)

	// Reset counts to zero to provide a predictable demo state.
	c.Store.Set("mutations", float64(0))
	c.Store.Set("intersections", float64(0))
	headerComponent := NewHeaderComponent(map[string]interface{}{
		"title": "Observer Component",
	})
	c.AddDependency("header", headerComponent)

	return c
}

func (c *ObserverComponent) Mount() {
	c.HTMLComponent.Mount()
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
	mutateBtn := jsa.Global().Get("document").Call("getElementById", "mutateBtn")
	mutateCh := events.Listen("click", mutateBtn)
	go func() {
		for range mutateCh {
			node := jsa.Global().Get("document").Call("getElementById", "observeMe")
			child := jsa.Global().Get("document").Call("createElement", "span")
			node.Call("appendChild", child)
		}
	}()

	opts := jst.ValueOf(map[string]any{})
	intCh, stopInt := events.ObserveIntersections(".watched", opts)
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
	toggleBtn := jsa.Global().Get("document").Call("getElementById", "toggleBtn")
	toggleCh := events.Listen("click", toggleBtn)
	go func() {
		for range toggleCh {
			el := jsa.Global().Get("document").Call("querySelector", ".watched")
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
