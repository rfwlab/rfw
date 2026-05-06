# State Management

rfw v2 provides two reactive state primitives: **signals** for local component state and **stores** for shared global state. Both are auto-wired by `composition.New` based on field types — no tags required.

---

## Signals

Signals are fine-grained reactive values. When a signal changes, only the template bindings and effects that read it re-render.

### Creation

Use the `t` package constructors:

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

Under the hood, `t.Int` is `*state.Signal[int]`, `t.String` is `*state.Signal[string]`, etc.

Signals are nil-safe: calling `.Get()` or `.Set()` on a nil `*Signal[T]` is a no-op (returns zero value for Get).

### Auto-Wiring by Type

Declare signal fields on your struct using value types or pointer types:

```go
type Counter struct {
    composition.Component
    Count t.Int      // value type
    Name  *t.String  // pointer type (auto-initialized if nil)
}

func (c *Counter) Inc() { c.Count.Set(c.Count.Get() + 1) }
func (c *Counter) Dec() { c.Count.Set(c.Count.Get() - 1) }
```

`composition.New` detects signal-type fields and:

- If the field is a **value type** (`t.Int`), registers it as a prop via `field.Addr()`.
- If the field is a **pointer type** and **nil** (`*t.Int`), auto-creates a zero-value signal and sets the field.
- If the field is a **pointer type** and **non-nil**, registers it directly as a prop.

Initialize with values:

```go
view, err := composition.New(&Counter{
    Count: *t.NewInt(5),
})
```

Or rely on auto-zero and set in `OnMount`:

```go
type Counter struct {
    composition.Component
    Count *t.Int
}

func (c *Counter) OnMount() {
    c.Count.Set(1)
}
```

### Template Binding

**Read-only**, displays current value, updates DOM on change:

```rtml
<span>@signal:Count</span>
<p>Name: @signal:Name</p>
```

**Two-way**, appends `:w` to write back from form controls:

```rtml
<input value="@signal:Name:w">
<input type="checkbox" checked="@signal:Done:w">
<textarea>@signal:Bio:w</textarea>
```

Input changes write to the signal, and all other `@signal:Name` bindings update automatically.

---

## Stores

Stores are namespaced key-value maps shared across components. They support modules, watchers, computed values, history (undo/redo), and persistence.

### Creation

```go
import "github.com/rfwlab/rfw/v2/state"

s := state.NewStore("cart", state.WithModule("app"))
s.Set("count", 0)
s.Set("total", 0.0)
```

Options:

| Option | Purpose |
|--------|---------|
| `state.WithModule("app")` | Namespace under a module |
| `state.WithHistory(10)` | Enable undo/redo with limit |
| `state.WithPersistence()` | Persist to localStorage |
| `state.WithDevTools()` | Log mutations (debug only) |

### Auto-Wiring with `*t.Store`

```go
type CartPage struct {
    composition.Component
    Cart *t.Store
}
```

`composition.New` detects `*t.Store` fields and calls `comp.Store("Cart")`, creating or retrieving the store scoped to the component.

### Template Binding

Stores are referenced by fully-qualified path: `@store:module.store.key`

```rtml
<p>Items: @store:app.cart.count</p>
<input value="@store:app.cart.name:w">
```

- Module defaults to `app`, so `@store:app.default.count` references the default store.
- Append `:w` for two-way binding.

---

## Computed Values

### @expr: Inline Computed

Use `@expr:` in templates for inline derived values that re-evaluate when dependencies change:

```rtml
<p>Double: @expr:Count.Get * 2</p>
<p>Greeting: @expr:Name.Get + " world"</p>
```

Supports arithmetic (`+`, `-`, `*`, `/`), comparisons (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical operators (`&&`, `||`, `!`), and field access (`.Get`).

### Store Computed Values

Define computed values on stores that re-evaluate when dependencies change:

```go
state.Map(s, "double", "count", func(v int) int { return v * 2 })

state.Map2(s, "fullName", "first", "last", func(first, last string) string {
    return first + " " + last
})
```

Or custom computed values:

```go
s.RegisterComputed(state.NewComputed(
    "profile",
    []string{"first", "last", "age"},
    func(m map[string]any) any {
        return fmt.Sprintf("%s %s (%d)", m["first"], m["last"], m["age"])
    },
))
```

Template reference:

```rtml
<p>Full name: @store:app.default.fullName</p>
```

---

## Watchers

React to store changes with watchers:

```go
s.Watch("age", func(v any) {
    log.Println("age updated", v)
})
```

Or via `RegisterWatcher` for multi-key observation:

```go
remove := s.RegisterWatcher(state.NewWatcher(
    []string{"first", "last"},
    func(m map[string]any) {
        log.Println(m["first"], m["last"])
    },
    state.WatcherImmediate(),
))
// later: remove()
```

- `WatcherDeep()`, match nested path changes
- `WatcherImmediate()`, trigger on registration

---

## Undo / Redo with `*t.History`

Use `*t.History` fields for undo/redo on stores:

```go
type Editor struct {
    composition.Component
    Doc  *t.Store
    Hist *t.History
}
```

`composition.New` discovers `*t.History` fields and binds them to the component's first store automatically.

Call methods in event handlers:

```go
func (e *Editor) Save()    { e.Hist.Snapshot() }
func (e *Editor) Undo()    { e.Hist.Undo() }
func (e *Editor) Redo()    { e.Hist.Redo() }
```

Template:

```rtml
<button @on:click:Undo>Undo</button>
<button @on:click:Redo>Redo</button>
```

You can also create a history manually:

```go
hist := t.NewHistory(50)  // max 50 snapshots
hist.Bind(myStore)
hist.Snapshot()
hist.Undo()
hist.Redo()
```

---

## Store Managers

By default, all stores are registered in the `GlobalStoreManager`. Create isolated managers for testing or sandboxing:

```go
sm := state.NewStoreManager()
s := sm.NewStore("test", state.WithModule("app"))
```

---

## Actions

Bundle mutations into reusable functions:

```go
s := state.NewStore("counter")
s.Set("count", 0)

increment := state.Action(func(ctx state.Context) error {
    current, _ := s.Get("count").(int)
    s.Set("count", current+1)
    return nil
})

_ = state.Dispatch(context.Background(), increment)
```

---

## Persistence

Enable localStorage persistence:

```go
s := state.NewStore("profile", state.WithModule("app"), state.WithPersistence())
s.Set("name", "Ada")
```

Values survive browser reloads.

---

## DI Injection with `*t.Inject[T]`

Auto-fill struct fields from the DI container:

```go
composition.Container().Provide("logger", &MyLogger{})

type Page struct {
    composition.Component
    Logger *t.Inject[Logger]
}
```

`composition.New` resolves the field from `Container().Get("logger")` and sets the inner `Value`. Works for any type — services, configs, API clients.

---

## Effects

`state.Effect` tracks dependencies and re-runs when signals change. Used internally by the framework for template bindings. Prefer lifecycle hooks (`OnMount`/`OnUnmount`) or watchers in application code:

```go
stop := state.Effect(func() func() {
    fmt.Println("count is", count.Get())
    return nil
})
// later: stop()
```

---

## Type Reference

### Signal Constructors

| Constructor | Type | Zero value |
|---|---|---|
| `t.NewInt(v)` | `*t.Int` | `0` |
| `t.NewString(v)` | `*t.String` | `""` |
| `t.NewBool(v)` | `*t.Bool` | `false` |
| `t.NewFloat(v)` | `*t.Float` | `0.0` |
| `t.NewAny(v)` | `*t.Any` | `nil` |

All support `.Get()` and `.Set()`. Nil-safe: calling `.Get()` on a nil pointer returns the zero value.

### Host Signal Types

| Type | Underlying | Use |
|---|---|---|
| `t.HInt` | `*Signal[int]` | SSC host-synced integer |
| `t.HString` | `*Signal[string]` | SSC host-synced string |
| `t.HBool` | `*Signal[bool]` | SSC host-synced boolean |
| `t.HFloat` | `*Signal[float64]` | SSC host-synced float |

### Special Types

| Type | Use |
|---|---|
| `*t.Store` | Component-scoped store |
| `*t.Ref` | Template DOM ref |
| `*t.Inject[T]` | DI injection |
| `*t.History` | Undo/redo bound to a store |
| `*t.View` | Child view (include slot) |
| `*t.Slice[T]` | Reactive slice signal |
| `*t.Map[K,V]` | Reactive map signal |
| `t.Prop[T]` | Reactive prop |

### Store Options

| Option | Effect |
|---|---|
| `state.WithModule(m)` | Namespace under module `m` |
| `state.WithHistory(n)` | Enable undo/redo, max `n` entries |
| `state.WithPersistence()` | Persist to localStorage |
| `state.WithDevTools()` | Log all mutations |

### Template Directive Summary

| Directive | Purpose |
|---|---|
| `@signal:Name` | Read signal value |
| `@signal:Name:w` | Two-way signal binding |
| `@store:m.s.k` | Read store key |
| `@store:m.s.k:w` | Two-way store binding |
| `@expr:` | Inline computed expression |
| `@on:event:handler` | DOM event → Go method |

---

## Related

- [Signals & Effects](../essentials/signals-effects-and-watchers)
- [Composition](../essentials/composition)
- [Store vs Signals](./store-vs-signals)