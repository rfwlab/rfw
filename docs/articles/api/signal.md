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
| `Get() T` | Returns current value. Tracks in `Effect` if called inside one. |
| `Set(value T)` | Updates value and re-runs dependent effects. |

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