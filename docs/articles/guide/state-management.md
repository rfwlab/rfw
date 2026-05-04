# State Management

rfw v2 provides two reactive state primitives, **signals** for local component state and **stores** for shared global state. Both are auto-wired by `composition.New` via struct tags, eliminating manual prop passing for the common case.

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

### Auto-Wiring with rfw:"signal"

Declare signal fields on your struct with the `rfw:"signal"` tag:

```go
type Counter struct {
    composition.Component
    Count *t.Int    `rfw:"signal"`
    Name  *t.String `rfw:"signal"`
}

func (c *Counter) Inc() { c.Count.Set(c.Count.Get() + 1) }
func (c *Counter) Dec() { c.Count.Set(c.Count.Get() - 1) }
```

`composition.New` auto-wires these:

- If the field is **non-nil** (initialized at construction), it's registered as a prop directly.
- If the field is **nil**, a zero-value signal is created and set on the field automatically.

Initialize with values:

```go
view := composition.New(&Counter{
    Count: t.NewInt(0),
})
```

Or rely on auto-zero and set in `OnMount`:

```go
type Counter struct {
    composition.Component
    Count *t.Int `rfw:"signal"`
}

func (c *Counter) OnMount() {
    c.Count.Set(1)
}

view := composition.New(&Counter{})
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

### Auto-Wiring with rfw:"store:name"

```go
type CartPage struct {
    composition.Component
    Cart *state.Store `rfw:"store:cart"`
}
```

`composition.New` calls `comp.Store("cart")`, creating or retrieving the store scoped to the component.

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

## Undo / Redo

Enable history on stores with `WithHistory`:

```go
s := state.NewStore("profile", state.WithModule("app"), state.WithHistory(5))
s.Set("age", 30)
s.Set("age", 31)
s.Undo() // age → 30
s.Redo() // age → 31
```

Auto-wire undo/redo via `rfw:"history:store:undo:redo"`:

```go
type EditorPage struct {
    composition.Component
    Doc *state.Store `rfw:"store:doc"`
    _   struct{}     `rfw:"history:doc:Undo:redo"`
}
```

Template:

```rtml
<button @on:click:Undo>Undo</button>
<button @on:click:redo>Redo</button>
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

## DI Injection with rfw:"inject"

Auto-fill struct fields from the DI container:

```go
composition.Container().Register("logger", &MyLogger{})

type Page struct {
    composition.Component
    Logger *MyLogger `rfw:"inject"`
}
```

`composition.New` resolves the field from `Container().Get("logger")` and sets it. Works for any type, services, configs, API clients.

Custom key:

```go
type Page struct {
    composition.Component
    Log *MyLogger `rfw:"inject:appLogger"`
}
```

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

## No More Manual Prop()

In v1, you passed state through `map[string]any{}` and called handlers with `dom.RegisterHandlerFunc`. In v2, `rfw:` tags handle everything:

```go
// v1
func New() *core.HTMLComponent {
    c := core.NewComponent("Counter", tpl, map[string]any{"count": 0})
    dom.RegisterHandlerFunc("increment", func() {
        c.Props["count"] = c.Props["count"].(int) + 1
    })
    return c
}

// v2
type Counter struct {
    composition.Component
    Count *t.Int `rfw:"signal"`
}

func (c *Counter) Inc() { c.Count.Set(c.Count.Get() + 1) }

view := composition.New(&Counter{Count: t.NewInt(0)})
```

Type-safe, auto-wired, no map casts.

---

## API Reference

### Signal Constructors

| Constructor | Type | Zero value |
|---|---|---|
| `t.NewInt(v)` | `*t.Int` | `0` |
| `t.NewString(v)` | `*t.String` | `""` |
| `t.NewBool(v)` | `*t.Bool` | `false` |
| `t.NewFloat(v)` | `*t.Float` | `0.0` |
| `t.NewAny(v)` | `*t.Any` | `nil` |

All support `.Get()` and `.Set()`.

### Store Options

| Option | Effect |
|---|---|
| `state.WithModule(m)` | Namespace under module `m` |
| `state.WithHistory(n)` | Enable undo/redo, max `n` entries |
| `state.WithPersistence()` | Persist to localStorage |
| `state.WithDevTools()` | Log all mutations |

### rfw Tags Summary

| Tag | Field Type | Effect |
|---|---|---|
| `rfw:"signal"` | `*t.Int`, `*t.String`, etc. | Auto-wire as reactive prop |
| `rfw:"store:name"` | `*state.Store` | Create/retrieve named store |
| `rfw:"inject"` | Any pointer | Resolve from DI container |
| `rfw:"inject:key"` | Any pointer | Resolve with custom key |
| `rfw:"include:slot"` | `*t.View` | Wire child into `@include:slot` |
| `rfw:"host:Name"` | `string` | Register host component binding |
| `rfw:"event:click:handler"` |, | Register DOM event → method |
| `rfw:"template:path"` | `struct{}` | Override template path |
| `rfw:"history:store:undo:redo"` |, | Wire undo/redo on store |
| `rfw:"prop:name"` | Any | Create reactive prop |

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
- [Store vs Signals](../guide/store-vs-signals)