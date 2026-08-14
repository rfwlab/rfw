//go:build js && wasm

package core

import (
	"errors"
	"html"

	"github.com/rfwlab/rfw/v2/dom"
	http "github.com/rfwlab/rfw/v2/http"
	"github.com/rfwlab/rfw/v2/state"
)

// Suspense renders a fallback while its render function reports pending work.
type Suspense struct {
	render   func() (string, error)
	fallback string
	id       string
	mounted  bool
	last     string
	pending  string
	stop     func()
}

var _ Component = (*Suspense)(nil)

// NewSuspense creates a Suspense component with the given render function and fallback HTML.
func NewSuspense(render func() (string, error), fallback string) *Suspense {
	return &Suspense{
		render:   render,
		fallback: fallback,
		id:       generateComponentID("Suspense", nil),
	}
}

// Render executes the render function and shows the fallback until it resolves.
func (s *Suspense) Render() string {
	s.last = s.renderHTML()
	return s.last
}

func (s *Suspense) renderHTML() string {
	content := s.fallback
	if s.render == nil {
		return `<root data-component-id="` + s.id + `">` + content + `</root>`
	}
	rendered, err := s.render()
	switch {
	case errors.Is(err, http.ErrPending), errors.Is(err, state.ErrResourcePending):
	case err != nil:
		content = html.EscapeString(err.Error())
	default:
		content = rendered
	}
	return `<root data-component-id="` + s.id + `">` + content + `</root>`
}

// Mount subscribes to every reactive value read by the render function.
func (s *Suspense) Mount() {
	s.mounted = true
	if s.stop != nil {
		s.stop()
	}
	s.stop = state.Effect(func() func() {
		next := s.renderHTML()
		if s.mounted && next != s.last {
			s.pending = next
			requestScheduledRender(renderJob{
				id:    s.id,
				depth: mountedComponentDepth(s.id),
				active: func() bool {
					return s.mounted
				},
				render: func() {
					if s.pending == s.last {
						return
					}
					s.last = s.pending
					dom.UpdateMountedDOM(s.id, s.last)
				},
			})
		}
		return nil
	})
}

// Unmount releases the reactive render subscription.
func (s *Suspense) Unmount() {
	s.mounted = false
	cancelComponentRender(s.id)
	if s.stop != nil {
		s.stop()
		s.stop = nil
	}
}

// OnMount is a no-op for Suspense.
func (s *Suspense) OnMount() {}

// OnUnmount is a no-op for Suspense.
func (s *Suspense) OnUnmount() {}

// GetName returns the component name.
func (s *Suspense) GetName() string { return "Suspense" }

// GetID returns this Suspense instance ID.
func (s *Suspense) GetID() string { return s.id }

// SetSlots is a no-op since Suspense does not use slots.
func (s *Suspense) SetSlots(map[string]any) {}

// IsMounted reports whether Suspense is mounted.
func (s *Suspense) IsMounted() bool { return s.mounted }

// OnParams is a no-op since Suspense does not consume route parameters.
func (s *Suspense) OnParams(map[string]string) {}
