//go:build js && wasm

package composition

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
	t "github.com/rfwlab/rfw/v2/types"
)

type (
	Int    = t.Int
	String = t.String
	Bool   = t.Bool
	Float  = t.Float
	Any    = t.Any
	Store  = t.Store
	View   = t.View
	Comp   = t.Comp
)

type (
	Slice[T any]    = t.Slice[T]
	Map[K comparable, V any] = t.Map[K, V]
	Ref              = t.Ref
	Prop[T any]      = t.Prop[T]
)

type Viewer = t.Viewer

var (
	NewInt    = t.NewInt
	NewString = t.NewString
	NewBool   = t.NewBool
	NewFloat  = t.NewFloat
	NewAny    = t.NewAny
	NewRef    = t.NewRef
)

func NewSlice[T any](v ...[]T) *t.Slice[T] { return t.NewSlice(v...) }
func NewMap[K comparable, V any](v ...map[K]V) *t.Map[K, V] { return t.NewMap(v...) }
func NewProp[T any](v T) *t.Prop[T]          { return t.NewProp(v) }

func SetDevMode(v bool) { core.SetDevMode(v) }

var (
	_ = (*state.Store)(nil)
	_ = (*core.HTMLComponent)(nil)
)