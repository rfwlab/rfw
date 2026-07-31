//go:build js && wasm

// Package types provides framework state and component value types.
package types

import (
	"syscall/js"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
)

type (
	// Int is an integer signal.
	Int = state.Signal[int]
	// String is a string signal.
	String = state.Signal[string]
	// Bool is a boolean signal.
	Bool = state.Signal[bool]
	// Float is a float64 signal.
	Float = state.Signal[float64]
	// Any is an untyped signal.
	Any = state.Signal[any]
	// Store aliases the state store.
	Store = state.Store
	// View aliases an HTML component.
	View = core.HTMLComponent
	// Comp aliases the component interface.
	Comp = core.Component
)

// Slice is a slice signal.
type Slice[T any] struct {
	*state.Signal[[]T]
}

// NewSlice creates a slice signal.
func NewSlice[T any](v ...[]T) *Slice[T] {
	var initial []T
	if len(v) > 0 {
		initial = v[0]
	}
	return &Slice[T]{Signal: state.NewSignal(initial)}
}

// Map is a map signal.
type Map[K comparable, V any] struct {
	*state.Signal[map[K]V]
}

// NewMap creates a map signal.
func NewMap[K comparable, V any](v ...map[K]V) *Map[K, V] {
	var initial map[K]V
	if len(v) > 0 {
		initial = v[0]
	}
	return &Map[K, V]{Signal: state.NewSignal(initial)}
}

// HInt is a host-backed integer signal.
type HInt struct {
	*state.Signal[int]
}

// HString is a host-backed string signal.
type HString struct {
	*state.Signal[string]
}

// HBool is a host-backed boolean signal.
type HBool struct {
	*state.Signal[bool]
}

// HFloat is a host-backed float signal.
type HFloat struct {
	*state.Signal[float64]
}

// HAny is a host-backed untyped signal.
type HAny struct {
	*state.Signal[any]
}

// HSlice is a host-backed slice signal.
type HSlice[T any] struct {
	*state.Signal[[]T]
}

// HMap is a host-backed map signal.
type HMap[K comparable, V any] struct {
	*state.Signal[map[K]V]
}

// Ref stores a DOM reference.
type Ref struct {
	node js.Value
}

// NewRef creates an empty DOM reference.
func NewRef() *Ref {
	return &Ref{node: js.Null()}
}

// Set updates the referenced node.
func (r *Ref) Set(v js.Value) { r.node = v }

// Get returns the referenced node.
func (r *Ref) Get() js.Value { return r.node }

// IsNil reports whether no node is referenced.
func (r *Ref) IsNil() bool { return r.node.IsNull() || r.node.IsUndefined() }

// Prop stores a component property.
type Prop[T any] struct {
	value T
}

// NewProp creates a component property.
func NewProp[T any](v T) *Prop[T] {
	return &Prop[T]{value: v}
}

// Get returns the property value.
func (p *Prop[T]) Get() T { return p.value }

// Set updates the property value.
func (p *Prop[T]) Set(v T) { p.value = v }

// NewInt creates an integer signal.
func NewInt(v int) *Int { return state.NewSignal(v) }

// NewString creates a string signal.
func NewString(v string) *String { return state.NewSignal(v) }

// NewBool creates a boolean signal.
func NewBool(v bool) *Bool { return state.NewSignal(v) }

// NewFloat creates a float signal.
func NewFloat(v float64) *Float { return state.NewSignal(v) }

// NewAny creates an untyped signal.
func NewAny(v any) *Any { return state.NewSignal(v) }

// Inject marks a dependency injection field.
type Inject[T any] struct {
	Value T
}

// History records store snapshots for undo and redo.
type History struct {
	store   *state.Store
	max     int
	cursor  int
	entries []map[string]any
}

// NewHistory creates a bounded store history.
func NewHistory(limit int) *History {
	return &History{max: limit, entries: make([]map[string]any, 0)}
}

// Bind attaches a store to the history.
func (h *History) Bind(s *state.Store) {
	h.store = s
}

// Undo restores the previous snapshot.
func (h *History) Undo() {
	if h.store == nil || h.cursor <= 0 {
		return
	}
	h.cursor--
	snap := h.entries[h.cursor]
	for k, v := range snap {
		h.store.Set(k, v)
	}
}

// Redo restores the next snapshot.
func (h *History) Redo() {
	if h.store == nil || h.cursor >= len(h.entries)-1 {
		return
	}
	h.cursor++
	snap := h.entries[h.cursor]
	for k, v := range snap {
		h.store.Set(k, v)
	}
}

// Snapshot records the current store state.
func (h *History) Snapshot() {
	if h.store == nil {
		return
	}
	snap := h.store.Snapshot()
	if snap == nil {
		snap = map[string]any{}
	}
	if h.cursor < len(h.entries)-1 {
		h.entries = h.entries[:h.cursor+1]
	}
	h.entries = append(h.entries, snap)
	if len(h.entries) > h.max {
		h.entries = h.entries[len(h.entries)-h.max:]
	}
	h.cursor = len(h.entries) - 1
}

// Viewer exposes a component view.
type Viewer interface {
	View() *View
}
