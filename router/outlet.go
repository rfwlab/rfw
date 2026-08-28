//go:build js && wasm

package router

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/internal/rendertrace"
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
	installOutletRepaint()
	return c
}

var (
	outletRepaintInstalled bool
	repainting             bool
)

// installOutletRepaint keeps the routed subtree alive across re-renders of the
// shell around it. A persistent root bound to a store re-renders whenever that
// store changes, and its fresh markup carries an empty outlet: without this
// the routed page would disappear on the first store write after a navigation.
// Registered once, when the first outlet is built.
func installOutletRepaint() {
	if outletRepaintInstalled {
		return
	}
	outletRepaintInstalled = true
	core.OnTemplate(func(componentID, _ string) {
		if liveOutlet == nil || currentComponent == nil {
			return
		}
		if componentID == currentComponent.GetID() {
			return
		}
		liveOutlet.repaint()
	})
}

// repaint re-renders the routed component when its markup is no longer inside
// the outlet.
func (o *Outlet) repaint() {
	if repainting || currentComponent == nil {
		return
	}
	root := dom.ComponentRoot(o.GetID())
	if root.IsNull() || root.IsUndefined() {
		return
	}
	if child := root.Query("[data-component-id='" + currentComponent.GetID() + "']"); !child.IsNull() && !child.IsUndefined() {
		return
	}
	repainting = true
	defer func() { repainting = false }()
	o.renderChild(currentComponent, rendertrace.Cause{Kind: "parent"})
	currentComponent.Mount()
}

// OnMount registers this outlet as the live navigation target. If a route
// resolved before the outlet appeared (root mounted after InitRouter), the
// pending component renders immediately.
func (o *Outlet) OnMount() {
	liveOutlet = o
	o.HTMLComponent.OnMount()
	if currentComponent != nil {
		o.renderChild(currentComponent, rendertrace.Cause{Kind: "router"})
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
// different component trees leaves stale nodes behind. The marker div stays
// in place as the anchor, so a re-render of the shell around it can recognise
// the subtree as the router's and leave it alone.
func (o *Outlet) renderChild(c core.Component, cause rendertrace.Cause) {
	root := dom.ComponentRoot(o.GetID())
	if root.IsNull() || root.IsUndefined() {
		renderComponent(c, cause, "", 0, func(html string) {
			dom.UpdateDOM(c.GetID(), html)
		})
		return
	}
	target := root.Query("[data-router-outlet]")
	if target.IsNull() || target.IsUndefined() {
		target = root
	}
	parentID, depth := "", 0
	if rendertrace.Enabled() {
		parentID = o.GetID()
		depth = componentDOMDepth(parentID) + 1
	}
	renderComponent(c, cause, parentID, depth, func(html string) {
		dom.UpdateDOMIn(target, c.GetID(), html)
	})
}

// mountedRoot pins the persistent root: without a live reference the GC
// finalizer would tear down a mounted component under the user.
var mountedRoot core.Component

// MountRoot renders a persistent root component into #app and mounts it. The
// root lives outside the navigation lifecycle: the router only ever touches
// the outlet inside it. Call it before InitRouter.
func MountRoot(c core.Component) {
	mountedRoot = c
	renderComponent(mountedRoot, rendertrace.Cause{Kind: "mount"}, "", 0, func(html string) {
		dom.UpdateDOM(mountedRoot.GetID(), html)
	})
	mountedRoot.Mount()
	core.TriggerMount(mountedRoot)
}

func renderComponent(c core.Component, cause rendertrace.Cause, parentID string, depth int, commit func(string)) {
	if !rendertrace.Enabled() {
		commit(core.TryRender(c))
		return
	}
	cause = rendertrace.NormalizeCause(cause)
	batchID := rendertrace.NextBatchID()
	renderID := rendertrace.NextRenderID()
	started := rendertrace.NowMS()
	base := rendertrace.Record{
		BatchID:           batchID,
		RenderID:          renderID,
		ComponentID:       c.GetID(),
		ComponentName:     c.GetName(),
		ParentComponentID: parentID,
		Depth:             depth,
		Cause:             cause,
		Causes:            []rendertrace.Cause{cause},
	}
	startedRecord := base
	startedRecord.Event = "started"
	rendertrace.Emit(startedRecord)

	var templateMS, domMS float64
	phase, phaseStarted := "template", rendertrace.NowMS()
	defer func() {
		if recovered := recover(); recovered != nil {
			now := rendertrace.NowMS()
			switch phase {
			case "template":
				templateMS = now - phaseStarted
			case "dom":
				domMS = now - phaseStarted
			}
			failed := base
			failed.Event = "failed"
			failed.TemplateMS = templateMS
			failed.DOMMS = domMS
			failed.TotalMS = rendertrace.NowMS() - started
			failed.Outcome = "failed"
			failed.Reason = "panic"
			rendertrace.Emit(failed)
			panic(recovered)
		}
	}()

	html := core.TryRender(c)
	templateMS = rendertrace.NowMS() - phaseStarted
	phase, phaseStarted = "dom", rendertrace.NowMS()
	commit(html)
	domMS = rendertrace.NowMS() - phaseStarted
	phase = ""

	completed := base
	completed.Event = "committed"
	completed.TemplateMS = templateMS
	completed.DOMMS = domMS
	completed.TotalMS = rendertrace.NowMS() - started
	completed.Outcome = "committed"
	rendertrace.Emit(completed)
}

func componentDOMDepth(componentID string) int {
	root := dom.ComponentRoot(componentID)
	if root.IsNull() || root.IsUndefined() || root.Attr("data-component-id") != componentID {
		return 0
	}
	depth := 0
	for parent := root.Get("parentElement"); parent.Truthy(); parent = parent.Get("parentElement") {
		if parent.Call("hasAttribute", "data-component-id").Bool() {
			depth++
		}
	}
	return depth
}
