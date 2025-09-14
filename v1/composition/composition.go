//go:build js && wasm

// Package composition provides helpers for embedding existing HTML components
// inside typed wrappers.
package composition

import (
	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/state"
)

const compositionModule = "composition"

// Component wraps a *core.HTMLComponent for composition purposes.
type Component struct {
	*core.HTMLComponent
	createdStores map[string]struct{}
}

type signalAny interface{ Read() any }

// Wrap returns a composition.Component around c.
// Panics if c is nil.
func Wrap(c *core.HTMLComponent) *Component {
	if c == nil {
		panic("composition.Wrap: nil HTMLComponent")
	}
	comp := &Component{HTMLComponent: c, createdStores: make(map[string]struct{})}
	c.SetComponent(comp)
	return comp
}

// Unwrap returns the underlying *core.HTMLComponent.
func (c *Component) Unwrap() *core.HTMLComponent { return c.HTMLComponent }

// On registers fn under name in the DOM handler registry.
func (c *Component) On(name string, fn func()) {
	if name == "" {
		panic("composition.On: empty name")
	}
	if fn == nil {
		panic("composition.On: nil handler")
	}
	dom.RegisterHandlerFunc(name, fn)
}

// Prop associates a reactive signal with the component under the provided key.
// The signal is stored in the "composition" state namespace and added to the
// component's props.
func (c *Component) Prop(key string, sig signalAny) {
	if key == "" {
		panic("composition.Prop: empty key")
	}
	if sig == nil {
		panic("composition.Prop: nil signal")
	}
	store := state.GlobalStoreManager.GetStore(compositionModule, c.ID)
	if store == nil {
		store = state.NewStore(c.ID, state.WithModule(compositionModule))
		state.GlobalStoreManager.RegisterStore(compositionModule, c.ID, store)
	}
	store.Set(key, sig)

	if c.HTMLComponent.Props == nil {
		c.HTMLComponent.Props = make(map[string]any)
		c.HTMLComponent.Props[key] = sig
		return
	}
	if v, ok := c.HTMLComponent.Props[key]; ok {
		if _, ok := v.(signalAny); ok {
			c.HTMLComponent.Props[key] = sig
		}
		return
	}
	c.HTMLComponent.Props[key] = sig
}

// OnUnmount cleans up the composition store when the component is removed.
func (c *Component) OnUnmount() {
	for name := range c.createdStores {
		state.GlobalStoreManager.UnregisterStore(c.ID, name)
	}
	state.GlobalStoreManager.UnregisterStore(compositionModule, c.ID)
	// DOM handlers registered with On or History remain in the global registry
	// because the dom package does not provide an unregister API yet.
	c.HTMLComponent.OnUnmount()
}

// FromProp retrieves a signal from props or creates one with a default value.
// If the prop exists and holds a signal, it is returned directly. If the prop
// holds a plain value matching T, the value is wrapped in a new signal. If the
// value type is incompatible, FromProp panics. When the prop is missing, a new
// signal seeded with def is created.
func FromProp[T any](c *Component, key string, def T) *state.Signal[T] {
	if key == "" {
		panic("composition.FromProp: empty key")
	}
	if c.HTMLComponent.Props != nil {
		if v, ok := c.HTMLComponent.Props[key]; ok {
			if s, ok := v.(*state.Signal[T]); ok {
				return s
			}
			if val, ok := v.(T); ok {
				s := state.NewSignal(val)
				c.Prop(key, s)
				return s
			}
			panic("composition.FromProp: incompatible prop type")
		}
	}
	s := state.NewSignal(def)
	c.Prop(key, s)
	return s
}
