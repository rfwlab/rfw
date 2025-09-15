# signal

Fine-grained reactive values for local state.

| Function | Description |
| --- | --- |
| `NewSignal(initial)` | Create a signal with an initial value. |
| `Get()` | Read the current value. |
| `Set(value)` | Update the value and notify dependents. |
| `Effect(fn)` | Run `fn` when accessed signals change and return a stop function. |

