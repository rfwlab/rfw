//go:build js && wasm

package core

import "github.com/rfwlab/rfw/v2/dom"

// ErrorBoundary wraps a child component and renders a fallback UI when the
// child panics during Render or Mount. Once a panic occurs, the fallback UI is
// displayed for subsequent renders.
type ErrorBoundary struct {
	Child    Component
	Fallback string
	err      any
	mounted  bool
}

var _ Component = (*ErrorBoundary)(nil)

// NewErrorBoundary creates a new ErrorBoundary around the provided child
// component. If the child panics during Render or Mount, the provided fallback
// HTML will be rendered instead.
func NewErrorBoundary(child Component, fallback string) *ErrorBoundary {
	return &ErrorBoundary{Child: child, Fallback: fallback}
}

func (e *ErrorBoundary) fallbackHTML() string {
	return "<root data-component-id=\"" + e.Child.GetID() + "\">" + e.Fallback + "</root>"
}

// Render renders the child component, returning the fallback HTML if the child
// panics or if a previous panic was recorded.
func (e *ErrorBoundary) Render() (out string) {
	if e.err != nil {
		return e.fallbackHTML()
	}
	defer func() {
		if r := recover(); r != nil {
			e.err = r
			ReportError(r, "Boundary render: "+e.Child.GetName())
			out = e.fallbackHTML()
		}
	}()
	return e.Child.Render()
}

// Mount mounts the child component, updating the DOM with the fallback HTML if
// the child panics during mounting.
func (e *ErrorBoundary) Mount() {
	if e.err != nil {
		e.mounted = true
		return
	}
	defer func() {
		if r := recover(); r != nil {
			e.err = r
			e.mounted = true
			ReportError(r, "Boundary mount: "+e.Child.GetName())
			dom.UpdateDOM(e.Child.GetID(), e.Fallback)
		}
	}()
	e.Child.Mount()
	e.mounted = true
}

// Unmount delegates to the child component's Unmount method.
func (e *ErrorBoundary) Unmount() {
	e.Child.Unmount()
	e.mounted = false
}

// OnMount is a no-op for ErrorBoundary.
func (e *ErrorBoundary) OnMount() {}

// OnUnmount is a no-op for ErrorBoundary.
func (e *ErrorBoundary) OnUnmount() {}

// GetName returns the name of the component.
func (e *ErrorBoundary) GetName() string { return "ErrorBoundary" }

// GetID returns the wrapped child's ID.
func (e *ErrorBoundary) GetID() string { return e.Child.GetID() }

// SetSlots delegates slot assignment to the child component.
func (e *ErrorBoundary) SetSlots(slots map[string]any) {
	if e.Child != nil {
		e.Child.SetSlots(slots)
	}
}

// IsMounted reports whether the boundary or its fallback is mounted.
func (e *ErrorBoundary) IsMounted() bool { return e.mounted }

// OnParams delegates route parameters to the wrapped component.
func (e *ErrorBoundary) OnParams(params map[string]string) {
	e.Child.OnParams(params)
}
