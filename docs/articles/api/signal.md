# signal

Fine-grained reactive values for local state.

| Function | Description |
| --- | --- |
| `NewSignal(initial)` | Create a signal with an initial value. |
| `Get()` | Read the current value. |
| `Set(value)` | Update the value and notify dependents. |
| `Effect(fn)` | Run `fn` when accessed signals change and return a stop function. |

For a practical introduction to signals and effects, see [Signals & Effects](../essentials/signals-and-effects).

## Usage

```go
count := state.NewSignal(0)
stop := state.Effect(func() func() {
    v := count.Get()
    fmt.Println("count is", v)
    return nil
})

defer stop()
count.Set(1) // triggers the effect
```

`Effect` automatically tracks which signals are read inside. When `Set` updates a signal, only effects that called `Get` on that signal re-run. If the function passed to `Effect` returns a cleanup callback, it runs before each re-execution and when the effect stops.
