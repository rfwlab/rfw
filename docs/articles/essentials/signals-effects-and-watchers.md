# Signals, Effects & Watchers

Signals are **fine-grained reactive values**, when a signal changes, only the computations and template bindings that read it re-run. They are ideal for local component state without global stores.

---

## Creating a Signal

Use the `t` (types) package constructors to create typed signals:

```go
import t "github.com/rfwlab/rfw/v2/types"

count := t.NewInt(0)
name  := t.NewString("")
done  := t.NewBool(false)
price := t.NewFloat(9.99)
data  := t.NewAny(nil)
```

Read with `.Get()`, write with `.Set()`:

```go
count.Set(42)
fmt.Println(count.Get()) // 42
```

The underlying type is `*t.Int` (alias for `*state.Signal[int]`), `*t.String` (alias for `*state.Signal[string]`), etc. Each signal tracks its own dependents and notifies only them on change.

---

## Wiring Signals with composition.New

When building a component, declare signal fields with the `rfw:"signal"` tag and `composition.New` auto-wires them:

```go
import (
    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

type Counter struct {
    composition.Component
    Count *t.Int    `rfw:"signal"`
    Name  *t.String `rfw:"signal"`
}

func main() {
    view := composition.New(&Counter{
        Count: t.NewInt(0),
        Name:  t.NewString("hello"),
    })
    _ = view
}
```

If a signal field is nil at construction time, `composition.New` creates a zero-value signal automatically. You can also initialize signals in `OnMount`:

```go
func (c *Counter) OnMount() {
    c.Count.Set(1)
}
```

---

## Using Signals in Templates

### Read-only binding

Use `@signal:Name` to display a signal's current value. The DOM updates whenever the signal changes:

```rtml
<p>Count: @signal:Count</p>
<p>Name: @signal:Name</p>
```

### Writable binding

Append `:w` to make form controls write back to the signal:

```rtml
<input value="@signal:Name:w">
<input type="checkbox" checked="@signal:Done:w">
<textarea>@signal:Bio:w</textarea>
```

Typing or toggling updates the signal, and every other binding to that signal updates automatically.

### Computed expressions with @expr:

Use `@expr:` for inline computed values. The expression is re-evaluated whenever any referenced signal changes:

```rtml
<p>Double: @expr:Count.Get * 2</p>
<p>Label: @expr:Name.Get + " world"</p>
```

The `@expr:` directive supports arithmetic (`+`, `-`, `*`, `/`), comparisons (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical operators (`&&`, `||`, `!`), and field access (`.Get`).

---

## Conditionals and Loops

Signals integrate with `@if:` and `@for:` directives:

```rtml
@if:signal:Count == "3"
  <p>Three!</p>
@endif

@for:item in signal:Items
  <li>{item.Text}</li>
@endfor
```

When `Items` holds a slice or map, changes patch only the affected DOM nodes.

---

## Effects (Internal)

The `state.Effect` function is used internally by the framework to track signal dependencies in templates and computed expressions. It is **not exposed** as a public API for application code. Use lifecycle hooks (`OnMount` / `OnUnmount`) or watchers instead.

---

## Passing Signals as Props

Provide signals through component props so children can bind to them:

```go
view := composition.New(&Child{
    Count: parentCount,
})
```

The child template can then use `@signal:Count` to stay reactive.

---

## API Reference

| Constructor    | Type         | Zero value |
| -------------- | ------------ | ---------- |
| `t.NewInt(v)`  | `*t.Int`     | `0`        |
| `t.NewString(v)` | `*t.String` | `""`       |
| `t.NewBool(v)` | `*t.Bool`    | `false`    |
| `t.NewFloat(v)` | `*t.Float`  | `0.0`      |
| `t.NewAny(v)`  | `*t.Any`     | `nil`      |

All signal types support `.Get()` and `.Set()`.

---

## Why Use Signals

* **Local state**, no global store needed for component data.
* **Fine-grained**, only dependents re-run on change.
* **Declarative templates**, signals bind naturally in RTML.
* **Two-way binding**, `:w` makes form inputs reactive.
* **Computed values**, `@expr:` derives values from signals inline.