# Reactivity Fundamentals

rfw updates the DOM only where state changed. No virtual DOM, no manual patching.

---

## Signals

Signals are standalone reactive values. Create them with typed constructors:

```go
import t "github.com/rfwlab/rfw/v2/types"

count := t.NewInt(0)
name := t.NewString("rfw")
active := t.NewBool(false)
```

`Get()` reads, `Set()` writes and notifies dependents:

```go
count.Set(count.Get() + 1)
```

Signals are nil-safe: calling `.Get()` or `.Set()` on a nil `*Signal[T]` is a no-op (Get returns zero value).

In components, declare signal fields by type — no tags required:

```go
type Counter struct {
    composition.Component
    Count t.Int      // value type
    Name  *t.String  // pointer type (auto-initialized if nil)
}
```

`composition.New(&Counter{})` detects signal-type fields and auto-wires them. Nil pointer fields get zero-value signals.

In RTML, bind signals with `@signal:name`:

```rtml
<p>@signal:Count</p>
<input value="@signal:Count:w">
```

Append `:w` for two-way binding on form controls.

---

## Host Signal Types (SSC)

For server-side computed values, use host signal types:

```go
type VisitPage struct {
    composition.Component
    Visit t.HInt // signal + host component binding
}
```

`t.HInt`, `t.HString`, `t.HBool`, `t.HFloat` are signals that also auto-register as host component bindings for SSC.

---

## Stores

Stores hold shared reactive key/value pairs scoped to a component. Declare them with `*t.Store` fields:

```go
type App struct {
    composition.Component
    CounterStore *t.Store
}
```

`composition.New` creates or retrieves the store automatically. Access it via `@store` in RTML:

```rtml
<p>@store:default.counter.count</p>
<input value="@store:default.counter.count:w">
```

Stores can also be created manually:

```go
func (a *App) OnMount() {
    s := a.CounterStore
    s.Set("count", 0)
}
```

---

## Template Directives

| Directive | Writable | Purpose |
| --- | --- | --- |
| `@signal:Name` | No | Read a signal prop |
| `@signal:Name:w` | Yes | Two-way bind to form control |
| `@store:module.store.key` | No | Read a store value |
| `@store:module.store.key:w` | Yes | Two-way bind to store |
| `@expr:expression` | No | Computed expression |
| `@on:event:handler` | - | DOM event to Go method |

---

## @expr: Computed Expressions

RTML supports inline expressions with `@expr:`:

```rtml
<p>@expr:Count.Get * 2</p>
<p>@expr:Count.Get > 0</p>
<p>@expr:Name.Get + "!"</p>
```

These re-evaluate whenever any referenced signal changes. Prefer `@expr:` for simple derivations; use Go methods for complex logic.

---

## Fine-grained Updates

When a signal or store value changes, rfw patches only the DOM nodes that depend on it. Unrelated nodes remain untouched. This scales to large applications without re-rendering the entire component tree.

---

## Effects

`state.Effect` runs a function and re-runs it when accessed signals change:

```go
state.Effect(func() func() {
    fmt.Println("count:", count.Get())
    return nil
})
```

Returns a stop function. Optional cleanup runs before each re-execution.

---

## Summary

- Use signal types (`t.Int`, `*t.String`) for local reactive state; `*t.Store` for component-scoped shared state.
- `@signal:Name:w` and `@store:...:w` enable two-way binding on inputs.
- `@expr:` handles inline computed values in templates.
- Changes propagate only to affected nodes.
- Host signal types (`t.HInt`, etc.) add SSC host bindings.
- All signal methods are nil-safe.