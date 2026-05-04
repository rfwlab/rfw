//go:build !js || !wasm

package types

import (
	"github.com/rfwlab/rfw/v2/core"
)

type signalStub[T any] struct {
	value T
}

func (s *signalStub[T]) Get() T  { return s.value }
func (s *signalStub[T]) Set(v T) { s.value = v }
func (s *signalStub[T]) Read() any { return s.value }

type (
	Int    = signalStub[int]
	String = signalStub[string]
	Bool   = signalStub[bool]
	Float  = signalStub[float64]
	Any    = signalStub[any]
	Store  = core.Component
	View   = core.HTMLComponent
	Comp   = core.Component
)

type Slice[T any] struct {
	*signalStub[[]T]
}

type Map[K comparable, V any] struct {
	*signalStub[map[K]V]
}

type Ref struct{}

type Prop[T any] struct {
	value T
}

func (p *Prop[T]) Get() T  { return p.value }
func (p *Prop[T]) Set(v T) { p.value = v }

func NewInt(v int) *Int       { return &Int{} }
func NewString(v string) *String { return &String{} }
func NewBool(v bool) *Bool    { return &Bool{} }
func NewFloat(v float64) *Float { return &Float{} }
func NewAny(v any) *Any       { return &Any{} }
func NewSlice[T any](v ...[]T) *Slice[T] {
	var initial []T
	if len(v) > 0 {
		initial = v[0]
	}
	return &Slice[T]{signalStub: &signalStub[[]T]{value: initial}}
}
func NewMap[K comparable, V any](v ...map[K]V) *Map[K, V] {
	var initial map[K]V
	if len(v) > 0 {
		initial = v[0]
	}
	return &Map[K, V]{signalStub: &signalStub[map[K]V]{value: initial}}
}
func NewRef() *Ref { return &Ref{} }
func NewProp[T any](v T) *Prop[T] { return &Prop[T]{value: v} }

type Viewer interface {
	View() *View
}