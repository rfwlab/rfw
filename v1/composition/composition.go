//go:build js && wasm

// Package composition provides helpers for embedding existing HTML components
// inside typed wrappers.
package composition

import core "github.com/rfwlab/rfw/v1/core"

// Component wraps a *core.HTMLComponent for composition purposes.
type Component struct {
	*core.HTMLComponent
}

// Wrap returns a composition.Component around c.
// Panics if c is nil.
func Wrap(c *core.HTMLComponent) *Component {
	if c == nil {
		panic("composition.Wrap: nil HTMLComponent")
	}
	return &Component{HTMLComponent: c}
}

// Unwrap returns the underlying *core.HTMLComponent.
func (c *Component) Unwrap() *core.HTMLComponent { return c.HTMLComponent }
