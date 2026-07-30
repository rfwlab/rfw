//go:build js && wasm

package core

import (
	"errors"

	http "github.com/rfwlab/rfw/v2/http"
)

// Suspense renders a fallback while the render function returns http.ErrPending.
type Suspense struct {
	render   func() (string, error)
	fallback string
	mounted  bool
}

var _ Component = (*Suspense)(nil)

// NewSuspense creates a Suspense component with the given render function and fallback HTML.
func NewSuspense(render func() (string, error), fallback string) *Suspense {
	return &Suspense{render: render, fallback: fallback}
}

// Render executes the render function and shows the fallback until it resolves.
func (s *Suspense) Render() string {
	if s.render == nil {
		return s.fallback
	}
	content, err := s.render()
	if err != nil {
		if errors.Is(err, http.ErrPending) {
			return s.fallback
		}
		return err.Error()
	}
	return content
}

// Mount marks the Suspense component as mounted.
func (s *Suspense) Mount() { s.mounted = true }

// Unmount marks the Suspense component as unmounted.
func (s *Suspense) Unmount() { s.mounted = false }

// OnMount is a no-op for Suspense.
func (s *Suspense) OnMount() {}

// OnUnmount is a no-op for Suspense.
func (s *Suspense) OnUnmount() {}

// GetName returns the component name.
func (s *Suspense) GetName() string { return "Suspense" }

// GetID returns an empty ID for Suspense.
func (s *Suspense) GetID() string { return "" }

// SetSlots is a no-op since Suspense does not use slots.
func (s *Suspense) SetSlots(map[string]any) {}

// IsMounted reports whether Suspense is mounted.
func (s *Suspense) IsMounted() bool { return s.mounted }

// OnParams is a no-op since Suspense does not consume route parameters.
func (s *Suspense) OnParams(map[string]string) {}
