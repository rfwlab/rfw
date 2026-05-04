# Reactivity Fundamentals

rfw updates the DOM only where state changed. No virtual DOM, no manual patching.

---

## Signals

Signals are standalone reactive values. Create them with typed constructors:

```go
count := composition.NewInt(0)
name := composition.NewString("rfw")
active := composition.NewBool(false)
```

`Get()` reads, `Set()` writes and notifies dependents:

```go
count.Set(count.Get() + 1)
```

In components, declare signal fields with the `rfw:"signal"` tag:

```go
type Counter struct {
    Count *composition.Int `rfw:"signal"`
}
```

`composition.New(&Counter{})` auto-creates a zero-value signal if the field is nil, and wires it as a prop.

In RTML, bind signals with `@signal:name`:

```rtml
<p>@signal:Count</p>
<input value="@signal:Count:w">
```

Append `:w` for two-way binding on form controls.

---

## Stores

Stores hold shared reactive key/value pairs scoped to a component. Declare them with `rfw:"store:name"`:

```go
type App struct {
    CounterStore *composition.Store `rfw:"store:counter"`
}
```

`composition.New` creates the store automatically. Access it via `@store` in RTML:

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
| `@expr:expression` | No | Computed expression (see below) |

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

- Use `rfw:"signal"` for local reactive state; `rfw:"store:name"` for component-scoped shared state.
- `@signal:Name:w` and `@store:...:w` enable two-way binding on inputs.
- `@expr:` handles inline computed values in templates.
- Changes propagate only to affected nodes.