//go:build js && wasm

package types

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
	"syscall/js"
)

type (
	Int    = state.Signal[int]
	String = state.Signal[string]
	Bool   = state.Signal[bool]
	Float  = state.Signal[float64]
	Any    = state.Signal[any]
	Store  = state.Store
	View   = core.HTMLComponent
	Comp   = core.Component
)

type Slice[T any] struct {
	*state.Signal[[]T]
}

func NewSlice[T any](v ...[]T) *Slice[T] {
	var initial []T
	if len(v) > 0 {
		initial = v[0]
	}
	return &Slice[T]{Signal: state.NewSignal(initial)}
}

type Map[K comparable, V any] struct {
	*state.Signal[map[K]V]
}

func NewMap[K comparable, V any](v ...map[K]V) *Map[K, V] {
	var initial map[K]V
	if len(v) > 0 {
		initial = v[0]
	}
	return &Map[K, V]{Signal: state.NewSignal(initial)}
}

type HInt struct {
	*state.Signal[int]
}

type HString struct {
	*state.Signal[string]
}

type HBool struct {
	*state.Signal[bool]
}

type HFloat struct {
	*state.Signal[float64]
}

type HAny struct {
	*state.Signal[any]
}

type HSlice[T any] struct {
	*state.Signal[[]T]
}

type HMap[K comparable, V any] struct {
	*state.Signal[map[K]V]
}

type Ref struct {
	node js.Value
}

func NewRef() *Ref {
	return &Ref{node: js.Null()}
}

func (r *Ref) Set(v js.Value) { r.node = v }
func (r *Ref) Get() js.Value  { return r.node }
func (r *Ref) IsNil() bool    { return r.node.IsNull() || r.node.IsUndefined() }

type Prop[T any] struct {
	value T
}

func NewProp[T any](v T) *Prop[T] {
	return &Prop[T]{value: v}
}

func (p *Prop[T]) Get() T  { return p.value }
func (p *Prop[T]) Set(v T) { p.value = v }

func NewInt(v int) *Int         { return state.NewSignal(v) }
func NewString(v string) *String { return state.NewSignal(v) }
func NewBool(v bool) *Bool      { return state.NewSignal(v) }
func NewFloat(v float64) *Float { return state.NewSignal(v) }
func NewAny(v any) *Any         { return state.NewSignal(v) }

type Inject[T any] struct {
	Value T
}

type History struct {
	store  *state.Store
	max    int
	cursor int
	entries []map[string]any
}

func NewHistory(max int) *History {
	return &History{max: max, entries: make([]map[string]any, 0)}
}

func (h *History) Bind(s *state.Store) {
	h.store = s
}

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

type Viewer interface {
	View() *View
}