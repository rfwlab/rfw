# types

```go
import "github.com/rfwlab/rfw/v2/types"
```

Type aliases and utilities for common reactive primitives. Shorthand for `state` signals.

## Pre-aliased Signal Types

| Type | Definition | Constructor |
| --- | --- | --- |
| `Int` | `Signal[int]` | `NewInt(v int) *Int` |
| `String` | `Signal[string]` | `NewString(v string) *String` |
| `Bool` | `Signal[bool]` | `NewBool(v bool) *Bool` |
| `Float` | `Signal[float64]` | `NewFloat(v float64) *Float` |
| `Any` | `Signal[any]` | `NewAny(v any) *Any` |
| `Store` | `state.Store` | - |

## View & Component Shorthands

| Type | Definition |
| --- | --- |
| `View` | `core.HTMLComponent` |
| `Comp` | `core.Component` |

## Slice

```go
type Slice[T any] struct { *Signal[[]T] }
func NewSlice[T any](v ...[]T) *Slice[T]
```

Reactive slice container with signal methods.

## Map

```go
type Map[K comparable, V any] struct { *Signal[map[K]V] }
func NewMap[K comparable, V any](v ...map[K]V) *Map[K, V]
```

Reactive map container with signal methods.

## H* Types (Host Signal)

Helpers for server-side computed signals:

| Type | Definition |
| --- | --- |
| `HInt` | `Signal[int]` |
| `HString` | `Signal[string]` |
| `HBool` | `Signal[bool]` |
| `HFloat` | `Signal[float64]` |
| `HAny` | `Signal[any]` |
| `HSlice[T]` | `Signal[[]T]` |
| `HMap[K,V]` | `Signal[map[K]V]` |

## Ref

```go
type Ref struct { node js.Value }
func NewRef() *Ref
func (r *Ref) Set(v js.Value)
func (r *Ref) Get() js.Value
func (r *Ref) IsNil() bool
```

DOM reference wrapper for template refs.

## Prop

```go
type Prop[T any] struct { value T }
func NewProp[T any](v T) *Prop[T]
func (p *Prop[T]) Get() T
func (p *Prop[T]) Set(v T)
```

Reactive prop container, useful for component props.

## History

```go
type History struct {
    store  *Store
    max    int
    cursor int
    entries []map[string]any
}
func NewHistory(max int) *History
func (h *History) Bind(s *Store)
func (h *History) Undo()
func (h *History) Redo()
func (h *History) Snapshot()
```

Undo/redo manager bound to a store.