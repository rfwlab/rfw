//go:build js && wasm

package router

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

// Outlet is a plain component that marks where routed components render.
// Include one anywhere in your tree (typically inside an app shell mounted
// with MountRoot) and navigation swaps only its subtree: everything around it
// keeps its DOM, delegated handlers, and state. With no live outlet the
// router falls back to replacing #app wholesale, the pre-outlet behavior.
type Outlet struct {
	*core.HTMLComponent
}

var liveOutlet *Outlet

var outletTpl = []byte(`<root><div data-router-outlet></div></root>`)

// NewOutlet builds the outlet component; mount it via a dependency include.
func NewOutlet() *Outlet {
	c := &Outlet{HTMLComponent: core.NewHTMLComponent("RouterOutlet", outletTpl, nil)}
	c.SetComponent(c)
	c.Init(nil)
	return c
}

// OnMount registers this outlet as the live navigation target. If a route
// resolved before the outlet appeared (root mounted after InitRouter), the
// pending component renders immediately.
func (o *Outlet) OnMount() {
	liveOutlet = o
	o.HTMLComponent.OnMount()
	if currentComponent != nil {
		o.renderChild(currentComponent)
		currentComponent.Mount()
	}
}

// OnUnmount clears the live outlet (the shell around it is going away).
func (o *Outlet) OnUnmount() {
	if liveOutlet == o {
		liveOutlet = nil
	}
	o.HTMLComponent.OnUnmount()
}

// renderChild replaces the outlet subtree with the routed component's render.
// Route swaps replace wholesale on purpose: positionally diffing two
// different component trees leaves stale nodes behind.
func (o *Outlet) renderChild(c core.Component) {
	root := dom.ComponentRoot(o.GetID())
	if root.IsNull() || root.IsUndefined() {
		dom.UpdateDOM(c.GetID(), core.TryRender(c))
		return
	}
	dom.UpdateDOMIn(root, c.GetID(), core.TryRender(c))
}

// mountedRoot pins the persistent root: without a live reference the GC
// finalizer would tear down a mounted component under the user.
var mountedRoot core.Component

// MountRoot renders a persistent root component into #app and mounts it. The
// root lives outside the navigation lifecycle: the router only ever touches
// the outlet inside it. Call it before InitRouter.
func MountRoot(c core.Component) {
	mountedRoot = c
	dom.UpdateDOM(c.GetID(), core.TryRender(c))
	c.Mount()
	core.TriggerMount(c)
}
