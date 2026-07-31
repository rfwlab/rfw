//go:build js && wasm

package composition

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/state"
	t "github.com/rfwlab/rfw/v2/types"
)

type (
	// Int is an integer signal.
	Int = t.Int
	// String is a string signal.
	String = t.String
	// Bool is a boolean signal.
	Bool = t.Bool
	// Float is a floating-point signal.
	Float = t.Float
	// Any is an untyped signal.
	Any = t.Any
	// Store is a reactive state store.
	Store = t.Store
	// View is a renderable component view.
	View = t.View
	// Comp is a component alias.
	Comp = t.Comp
)

type (
	// Slice is a reactive slice.
	Slice[T any] = t.Slice[T]
	// Map is a reactive map.
	Map[K comparable, V any] = t.Map[K, V]
	// Ref holds a DOM element reference.
	Ref = t.Ref
	// Prop holds a component property.
	Prop[T any] = t.Prop[T]
	// HInt is a host-backed integer signal.
	HInt = t.HInt
	// HString is a host-backed string signal.
	HString = t.HString
	// HBool is a host-backed boolean signal.
	HBool = t.HBool
	// HFloat is a host-backed floating-point signal.
	HFloat = t.HFloat
	// HAny is an untyped host-backed signal.
	HAny = t.HAny
	// HSlice is a host-backed slice signal.
	HSlice[T any] = t.HSlice[T]
	// HMap is a host-backed map signal.
	HMap[K comparable, V any] = t.HMap[K, V]
)

// Viewer is implemented by values that expose a view.
type Viewer = t.Viewer

var (
	// NewInt creates an integer signal.
	NewInt = t.NewInt
	// NewString creates a string signal.
	NewString = t.NewString
	// NewBool creates a boolean signal.
	NewBool = t.NewBool
	// NewFloat creates a floating-point signal.
	NewFloat = t.NewFloat
	// NewAny creates an untyped signal.
	NewAny = t.NewAny
	// NewRef creates a DOM reference.
	NewRef = t.NewRef
)

// NewSlice creates a reactive slice.
func NewSlice[T any](v ...[]T) *t.Slice[T] { return t.NewSlice(v...) }

// NewMap creates a reactive map.
func NewMap[K comparable, V any](v ...map[K]V) *t.Map[K, V] { return t.NewMap(v...) }

// NewProp creates a component property.
func NewProp[T any](v T) *t.Prop[T] { return t.NewProp(v) }

// SetDevMode enables or disables development diagnostics.
func SetDevMode(v bool) { core.SetDevMode(v) }

var (
	_ = (*state.Store)(nil)
	_ = (*core.HTMLComponent)(nil)
)
