# signal

```go
import "github.com/rfwlab/rfw/v2/state"
```

Fine-grained reactive values. Re-exported from `composition/types` via the `state` package.

## NewSignal

```go
func NewSignal[T any](initial T) *Signal[T]
```

Creates a signal with the given initial value.

## Signal[T]

| Method | Description |
| --- | --- |
| `Get() T` | Returns current value. Tracks in `Effect` if called inside one. Nil-safe: returns zero value on nil receiver. |
| `Set(value T)` | Updates value and re-runs dependent effects. Nil-safe: no-op on nil receiver. |
| `Read() any` | Returns current value as `any`. Nil-safe: returns nil on nil receiver. |
| `OnChange(fn func(T)) func()` | Registers a change listener. Returns unsubscribe function. Nil-safe: returns no-op on nil receiver. |
| `Channel() <-chan T` | Returns a channel that receives new values. Nil-safe: returns nil channel on nil receiver. |
| `SetFromHost(v any)` | Updates value from a host (server-side) payload. Handles JSON float64→int conversion. |

## Nil Safety

All `Signal[T]` methods handle nil receivers gracefully — no panics. This is important when signals are accessed before `composition.New` initializes them, or in server-side rendering contexts.

## Effect

```go
func Effect(fn func() func()) func()
```

Re-runs `fn` whenever any signal read inside it changes. Returns a stop function.

## Pre-aliased Types

| Type | Definition | Constructor |
| --- | --- | --- |
| `Int` | `Signal[int]` | `NewInt(v int) *Int` |
| `String` | `Signal[string]` | `NewString(v string) *String` |
| `Bool` | `Signal[bool]` | `NewBool(v bool) *Bool` |
| `Float` | `Signal[float64]` | `NewFloat(v float64) *Float` |
| `Any` | `Signal[any]` | `NewAny(v any) *Any` |

## Host Signal Types

| Type | Definition | Use |
| --- | --- | --- |
| `HInt` | `*Signal[int]` | SSC host-synced integer |
| `HString` | `*Signal[string]` | SSC host-synced string |
| `HBool` | `*Signal[bool]` | SSC host-synced boolean |
| `HFloat` | `*Signal[float64]` | SSC host-synced float |
| `HSlice[T]` | `*Signal[[]T]` | SSC host-synced slice |
| `HMap[K,V]` | `*Signal[map[K]V]` | SSC host-synced map |

Host signal types embed `*Signal[T]` and additionally register as host component bindings in `composition.New`.