//go:build js && wasm

package core

import (
	"strings"
	"time"

	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
)

// Portal renders a child into a DOM target outside its component tree.
type Portal struct {
	id        string
	selector  string
	child     Component
	container dom.Element
	mounted   bool
}

// NewPortal creates a portal targeting a CSS selector.
func NewPortal(selector string, child Component) *Portal {
	return &Portal{
		id:       generateComponentID("Portal", map[string]any{"target": selector}),
		selector: selector,
		child:    child,
	}
}

func (portal *Portal) Render() string {
	return `<root data-component-id="` + portal.id + `"><template data-portal-anchor></template></root>`
}

func (portal *Portal) Mount() {
	if portal.mounted || portal.child == nil {
		return
	}
	target := dom.Query(portal.selector)
	if target.IsNull() || target.IsUndefined() {
		return
	}
	container := dom.CreateElement("div")
	container.SetAttr("data-portal-id", portal.id)
	target.AppendChild(container)
	portal.container = container
	dom.UpdateDOMIn(container, portal.child.GetID(), TryRender(portal.child))
	portal.child.Mount()
	portal.mounted = true
}

func (portal *Portal) Unmount() {
	if !portal.mounted {
		return
	}
	portal.child.Unmount()
	if !portal.container.IsNull() && !portal.container.IsUndefined() {
		portal.container.Call("remove")
	}
	portal.container = dom.Element{}
	portal.mounted = false
}

func (portal *Portal) OnMount()   {}
func (portal *Portal) OnUnmount() {}
func (portal *Portal) GetName() string {
	return "Portal"
}
func (portal *Portal) GetID() string { return portal.id }
func (portal *Portal) SetSlots(slots map[string]any) {
	if portal.child != nil {
		portal.child.SetSlots(slots)
	}
}
func (portal *Portal) IsMounted() bool { return portal.mounted }
func (portal *Portal) OnParams(params map[string]string) {
	if portal.child != nil {
		portal.child.OnParams(params)
	}
}

// KeepAliveAware receives activation events without being unmounted.
type KeepAliveAware interface {
	OnActivate()
	OnDeactivate()
}

// KeepAlive preserves a child's DOM and component state across route swaps.
type KeepAlive struct {
	id          string
	child       Component
	fragment    js.Value
	initialized bool
	cached      bool
	mounted     bool
	disposed    bool
}

// NewKeepAlive creates a state-preserving component wrapper.
func NewKeepAlive(child Component) *KeepAlive {
	return &KeepAlive{id: generateComponentID("KeepAlive", nil), child: child}
}

func (keep *KeepAlive) Render() string {
	content := ""
	if !keep.cached && !keep.disposed && keep.child != nil {
		content = TryRender(keep.child)
	}
	return `<root data-component-id="` + keep.id + `"><div data-keepalive-host>` + content + `</div></root>`
}

func (keep *KeepAlive) Mount() {
	if keep.mounted || keep.disposed || keep.child == nil {
		return
	}
	if keep.cached && keep.fragment.Truthy() {
		root := dom.ComponentRoot(keep.id)
		host := root.Query("[data-keepalive-host]")
		if !host.IsNull() && !host.IsUndefined() {
			host.Call("appendChild", keep.fragment)
		}
		keep.cached = false
	}
	if !keep.initialized {
		keep.child.Mount()
		keep.initialized = true
	} else if aware, ok := keep.child.(KeepAliveAware); ok {
		aware.OnActivate()
	}
	keep.mounted = true
}

func (keep *KeepAlive) Unmount() {
	if !keep.mounted || keep.disposed || keep.child == nil {
		return
	}
	root := dom.ComponentRoot(keep.id)
	if root.Attr("data-component-id") == keep.id {
		childRoot := root.Query(`[data-component-id="` + keep.child.GetID() + `"]`)
		if !childRoot.IsNull() && !childRoot.IsUndefined() {
			fragment := js.Document().Call("createDocumentFragment")
			fragment.Call("appendChild", childRoot.Value)
			keep.fragment = fragment
			keep.cached = true
		}
	}
	if aware, ok := keep.child.(KeepAliveAware); ok {
		aware.OnDeactivate()
	}
	keep.mounted = false
}

// Dispose permanently unmounts the cached child.
func (keep *KeepAlive) Dispose() {
	if keep.disposed {
		return
	}
	root := dom.ComponentRoot(keep.id)
	if keep.cached && keep.fragment.Truthy() {
		holder := dom.CreateElement("div")
		holder.SetStyle("display", "none")
		dom.Doc().Body().AppendChild(holder)
		holder.Call("appendChild", keep.fragment)
		keep.child.Unmount()
		holder.Call("remove")
	} else if keep.initialized {
		keep.child.Unmount()
	}
	keep.fragment = js.Undefined()
	keep.cached = false
	keep.mounted = false
	keep.disposed = true
	if root.Attr("data-component-id") == keep.id {
		root.Call("remove")
	}
}

func (keep *KeepAlive) OnMount()   {}
func (keep *KeepAlive) OnUnmount() {}
func (keep *KeepAlive) GetName() string {
	return "KeepAlive"
}
func (keep *KeepAlive) GetID() string { return keep.id }
func (keep *KeepAlive) SetSlots(slots map[string]any) {
	if keep.child != nil {
		keep.child.SetSlots(slots)
	}
}
func (keep *KeepAlive) IsMounted() bool { return keep.mounted }
func (keep *KeepAlive) OnParams(params map[string]string) {
	if keep.child != nil {
		keep.child.OnParams(params)
	}
}

// TransitionConfig names the CSS classes used during enter and leave phases.
type TransitionConfig struct {
	EnterFrom   string
	EnterActive string
	EnterTo     string
	LeaveFrom   string
	LeaveActive string
	LeaveTo     string
	Duration    time.Duration
}

// DefaultTransitionConfig returns class names compatible with plain CSS.
func DefaultTransitionConfig() TransitionConfig {
	return TransitionConfig{
		EnterFrom:   "rfw-enter-from",
		EnterActive: "rfw-enter-active",
		EnterTo:     "rfw-enter-to",
		LeaveFrom:   "rfw-leave-from",
		LeaveActive: "rfw-leave-active",
		LeaveTo:     "rfw-leave-to",
		Duration:    200 * time.Millisecond,
	}
}

// Transition applies CSS enter and leave phases around a child component.
type Transition struct {
	id      string
	child   Component
	config  TransitionConfig
	mounted bool
	timer   *time.Timer
	leaving dom.Element
}

// NewTransition creates a CSS transition wrapper.
func NewTransition(child Component, config TransitionConfig) *Transition {
	defaults := DefaultTransitionConfig()
	if config.EnterFrom == "" {
		config.EnterFrom = defaults.EnterFrom
	}
	if config.EnterActive == "" {
		config.EnterActive = defaults.EnterActive
	}
	if config.EnterTo == "" {
		config.EnterTo = defaults.EnterTo
	}
	if config.LeaveFrom == "" {
		config.LeaveFrom = defaults.LeaveFrom
	}
	if config.LeaveActive == "" {
		config.LeaveActive = defaults.LeaveActive
	}
	if config.LeaveTo == "" {
		config.LeaveTo = defaults.LeaveTo
	}
	if config.Duration == 0 {
		config.Duration = defaults.Duration
	}
	return &Transition{
		id:     generateComponentID("Transition", nil),
		child:  child,
		config: config,
	}
}

func (transition *Transition) Render() string {
	content := ""
	if transition.child != nil {
		content = TryRender(transition.child)
	}
	return `<root data-component-id="` + transition.id + `" data-transition>` + content + `</root>`
}

func (transition *Transition) Mount() {
	if transition.mounted || transition.child == nil {
		return
	}
	if transition.timer != nil {
		transition.timer.Stop()
		transition.timer = nil
	}
	if transition.leaving.Attr("data-component-id") == transition.id {
		transition.child.Unmount()
		transition.leaving.Call("remove")
		transition.leaving = dom.Element{}
	}
	transition.mounted = true
	transition.child.Mount()
	root := dom.ComponentRoot(transition.id)
	addClasses(root, transition.config.EnterFrom, transition.config.EnterActive)
	js.OnAnimationFrame(func() {
		if !transition.mounted {
			return
		}
		removeClasses(root, transition.config.EnterFrom)
		addClasses(root, transition.config.EnterTo)
	})
	transition.timer = time.AfterFunc(transition.config.Duration, func() {
		if transition.mounted {
			removeClasses(root, transition.config.EnterActive, transition.config.EnterTo)
		}
	})
}

func (transition *Transition) Unmount() {
	if !transition.mounted || transition.child == nil {
		return
	}
	transition.mounted = false
	if transition.timer != nil {
		transition.timer.Stop()
	}
	root := dom.ComponentRoot(transition.id)
	if root.Attr("data-component-id") != transition.id {
		transition.child.Unmount()
		return
	}
	dom.Doc().Body().AppendChild(root)
	transition.leaving = root
	removeClasses(root, transition.config.EnterFrom, transition.config.EnterActive, transition.config.EnterTo)
	addClasses(root, transition.config.LeaveFrom, transition.config.LeaveActive)
	js.OnAnimationFrame(func() {
		removeClasses(root, transition.config.LeaveFrom)
		addClasses(root, transition.config.LeaveTo)
	})
	transition.timer = time.AfterFunc(transition.config.Duration, func() {
		transition.child.Unmount()
		root.Call("remove")
		removeClasses(root, transition.config.LeaveActive, transition.config.LeaveTo)
		transition.leaving = dom.Element{}
		transition.timer = nil
	})
}

// Dispose removes a transition immediately.
func (transition *Transition) Dispose() {
	if transition.timer != nil {
		transition.timer.Stop()
		transition.timer = nil
	}
	if transition.child != nil && transition.child.IsMounted() {
		transition.child.Unmount()
	}
	root := dom.ComponentRoot(transition.id)
	if root.Attr("data-component-id") == transition.id {
		root.Call("remove")
	}
	transition.leaving = dom.Element{}
	transition.mounted = false
}

func (transition *Transition) OnMount()   {}
func (transition *Transition) OnUnmount() {}
func (transition *Transition) GetName() string {
	return "Transition"
}
func (transition *Transition) GetID() string { return transition.id }
func (transition *Transition) SetSlots(slots map[string]any) {
	if transition.child != nil {
		transition.child.SetSlots(slots)
	}
}
func (transition *Transition) IsMounted() bool { return transition.mounted }
func (transition *Transition) OnParams(params map[string]string) {
	if transition.child != nil {
		transition.child.OnParams(params)
	}
}

func addClasses(element dom.Element, groups ...string) {
	for _, group := range groups {
		for _, className := range strings.Fields(group) {
			element.AddClass(className)
		}
	}
}

func removeClasses(element dom.Element, groups ...string) {
	for _, group := range groups {
		for _, className := range strings.Fields(group) {
			element.RemoveClass(className)
		}
	}
}
