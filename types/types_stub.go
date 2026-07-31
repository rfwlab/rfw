//go:build !js || !wasm

// Package types provides framework state and component value types.
package types

import (
	"sync"

	"github.com/rfwlab/rfw/v2/core"
)

type signalStub[T any] struct {
	value T

	onChangeMu sync.Mutex
	onChange   []func(T)
	ch         chan T
	chCreated  bool
}

func (s *signalStub[T]) Get() T    { return s.value }
func (s *signalStub[T]) Set(v T)   { s.value = v; s.notifyOnChange(v) }
func (s *signalStub[T]) Read() any { return s.value }
func (s *signalStub[T]) SetFromHost(raw any) {
	if v, ok := raw.(T); ok {
		s.Set(v)
		return
	}
	switch any(s.value).(type) {
	case int:
		if f, ok := raw.(float64); ok {
			s.Set(any(int(f)).(T))
		}
	case float64:
		if f, ok := raw.(float64); ok {
			s.Set(any(f).(T))
		}
	case string:
		if str, ok := raw.(string); ok {
			s.Set(any(str).(T))
		}
	case bool:
		if b, ok := raw.(bool); ok {
			s.Set(any(b).(T))
		}
	}
}

func (s *signalStub[T]) OnChange(fn func(T)) *Subscription {
	s.onChangeMu.Lock()
	idx := len(s.onChange)
	s.onChange = append(s.onChange, fn)
	s.onChangeMu.Unlock()
	sub := &Subscription{
		cancel: func() {
			s.onChangeMu.Lock()
			defer s.onChangeMu.Unlock()
			if idx < len(s.onChange) {
				s.onChange[idx] = nil
			}
			s.maybeCloseChannel()
		},
	}
	return sub
}

func (s *signalStub[T]) Channel() <-chan T {
	s.onChangeMu.Lock()
	defer s.onChangeMu.Unlock()
	if !s.chCreated {
		s.ch = make(chan T, 8)
		s.chCreated = true
	}
	return s.ch
}

func (s *signalStub[T]) notifyOnChange(v T) {
	s.onChangeMu.Lock()
	listeners := make([]func(T), len(s.onChange))
	copy(listeners, s.onChange)
	ch := s.ch
	hasCh := s.chCreated
	s.onChangeMu.Unlock()

	for _, fn := range listeners {
		if fn != nil {
			fn(v)
		}
	}
	if hasCh && ch != nil {
		select {
		case ch <- v:
		default:
		}
	}
}

func (s *signalStub[T]) maybeCloseChannel() {
	allNil := true
	for _, fn := range s.onChange {
		if fn != nil {
			allNil = false
			break
		}
	}
	if allNil && s.chCreated {
		close(s.ch)
		s.ch = nil
		s.chCreated = false
		s.onChange = nil
	}
}

// Subscription represents a cancellable listener.
type Subscription struct {
	cancel func()
	once   sync.Once
}

// Stop removes the listener.
func (s *Subscription) Stop() { s.once.Do(s.cancel) }

type (
	// Int is an integer signal.
	Int = signalStub[int]
	// String is a string signal.
	String = signalStub[string]
	// Bool is a boolean signal.
	Bool = signalStub[bool]
	// Float is a float64 signal.
	Float = signalStub[float64]
	// Any is an untyped signal.
	Any = signalStub[any]
	// Store aliases the non-WASM store placeholder.
	Store = core.Component
	// View aliases the non-WASM component view.
	View = core.HTMLComponent
	// Comp aliases the non-WASM component interface.
	Comp = core.Component
)

// Slice is a slice signal.
type Slice[T any] struct {
	*signalStub[[]T]
}

// Map is a map signal.
type Map[K comparable, V any] struct {
	*signalStub[map[K]V]
}

// HInt is a host-backed integer signal placeholder.
type HInt struct {
	*signalStub[int]
}

// HString is a host-backed string signal placeholder.
type HString struct {
	*signalStub[string]
}

// HBool is a host-backed boolean signal placeholder.
type HBool struct {
	*signalStub[bool]
}

// HFloat is a host-backed float signal placeholder.
type HFloat struct {
	*signalStub[float64]
}

// HAny is a host-backed untyped signal placeholder.
type HAny struct {
	*signalStub[any]
}

// HSlice is a host-backed slice signal placeholder.
type HSlice[T any] struct {
	*signalStub[[]T]
}

// HMap is a host-backed map signal placeholder.
type HMap[K comparable, V any] struct {
	*signalStub[map[K]V]
}

// Ref is a DOM reference placeholder.
type Ref struct{}

// Prop stores a component property.
type Prop[T any] struct {
	value T
}

// Get returns the property value.
func (p *Prop[T]) Get() T { return p.value }

// Set updates the property value.
func (p *Prop[T]) Set(v T) { p.value = v }

// Inject marks a dependency injection field.
type Inject[T any] struct {
	Value T
}

// History is a non-WASM history placeholder.
type History struct {
	max int
}

// NewHistory creates a history placeholder.
func NewHistory(limit int) *History { return &History{max: limit} }

// Undo performs no work outside WASM.
func (h *History) Undo() {}

// Redo performs no work outside WASM.
func (h *History) Redo() {}

// Snapshot performs no work outside WASM.
func (h *History) Snapshot() {}

// NewInt creates an integer signal.
func NewInt(v int) *Int { return &signalStub[int]{value: v} }

// NewString creates a string signal.
func NewString(v string) *String { return &signalStub[string]{value: v} }

// NewBool creates a boolean signal.
func NewBool(v bool) *Bool { return &signalStub[bool]{value: v} }

// NewFloat creates a float signal.
func NewFloat(v float64) *Float { return &signalStub[float64]{value: v} }

// NewAny creates an untyped signal.
func NewAny(v any) *Any { return &signalStub[any]{value: v} }

// NewSlice creates a slice signal.
func NewSlice[T any](v ...[]T) *Slice[T] {
	var initial []T
	if len(v) > 0 {
		initial = v[0]
	}
	return &Slice[T]{signalStub: &signalStub[[]T]{value: initial}}
}

// NewMap creates a map signal.
func NewMap[K comparable, V any](v ...map[K]V) *Map[K, V] {
	var initial map[K]V
	if len(v) > 0 {
		initial = v[0]
	}
	return &Map[K, V]{signalStub: &signalStub[map[K]V]{value: initial}}
}

// NewRef creates a DOM reference placeholder.
func NewRef() *Ref { return &Ref{} }

// NewProp creates a component property.
func NewProp[T any](v T) *Prop[T] { return &Prop[T]{value: v} }

// Viewer exposes a component view.
type Viewer interface {
	View() *View
}
