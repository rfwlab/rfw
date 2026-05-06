# Global Store vs Local Signals

State in **rfw v2** is managed with two complementary tools: **global stores** and **local signals**. Stores are centralized containers for shared, persistent data. Signals are lightweight reactive variables scoped to a component. This guide covers the differences and when to use each.

---

## Local Signals

A signal is a reactive variable scoped to the component that creates it. In v2, use `types.NewInt`, `types.NewString`, etc. or `state.NewSignal[T]` directly. Signals are nil-safe: calling `.Get()` or `.Set()` on a nil pointer is a no-op.

### With `composition.New` and type detection

```go
type Counter struct {
    composition.Component

    Count types.Int      // value type
    Name  *types.String // pointer type (auto-initialized if nil)
}
```

`composition.New` detects signal-type fields and auto-creates nil pointers, registers as props, and wires reactivity. No struct tags needed.

### Manual creation

```go
count := types.NewInt(0)

count.Set(1)
fmt.Println(count.Get()) // 1
```

### Effects

```go
stop := state.Effect(func() func() {
    fmt.Println("count:", count.Get())
    return nil
})
defer stop()

count.Set(2)  // re-runs the effect
```

Every `Set` re-triggers only the effects that read the signal.

---

## Global Store

A store holds data shared across your application. Keys are namespaced by module.

### With `*types.Store` field

```go
type Profile struct {
    composition.Component

    Settings *types.Store
}
```

`composition.New` detects `*types.Store` fields and calls `comp.Store("Settings")`, creating or retrieving the store scoped to the component.

### Manual creation

```go
var profile = state.NewStore("profile", state.WithModule("user"))

func init() {
    profile.Set("first", "Ada")
    profile.Set("last", "Lovelace")
}
```

In templates, `@user/profile.first` binds directly. Updates propagate automatically. Stores also support computed values, watchers, persistence, and undo/redo.

---

## Type Reference

| Field Type | Detection | Auto-wiring |
|---|---|---|
| `types.Int`, `types.String`, `types.Bool`, `types.Float` | Signal value type | Register as reactive prop |
| `*types.Int`, `*types.String`, etc. | Signal pointer type | Auto-init if nil, register as prop |
| `types.HInt`, `types.HString`, etc. | Host signal type | Register as prop + host component |
| `*types.Store` | Store type | Create/retrieve scoped store |
| `*types.Ref` | Ref type | Allocate + resolve DOM on mount |
| `*types.Inject[T]` | DI inject type | Resolve from container |
| `*types.History` | History type | Bind to component store for undo/redo |

---

## Choosing Between Them

| Scenario | Store | Signal |
|----------|-------|--------|
| Data shared across many components | Yes | No |
| Temporary state in a single component | No | Yes |
| Persistence in `localStorage` | Yes (`WithPersistence`) | No |
| Fine-grained reactive updates | Depends | Yes |
| Undo/redo support | Yes (`WithHistory` or `*t.History`) | No |
| Server-side `h:` bindings | Yes | Host types only |

- **Stores** simplify synchronization of complex data across the app
- **Signals** shine for small, isolated pieces of state

---

## Undo/Redo with `*types.History`

Bind a history to a store for undo/redo:

```go
type Editor struct {
    composition.Component

    Doc  *types.Store
    Hist *types.History
}

func (e *Editor) Save()    { e.Hist.Snapshot() }
func (e *Editor) Undo()    { e.Hist.Undo() }
func (e *Editor) Redo()    { e.Hist.Redo() }
```

`composition.New` auto-binds `*t.History` fields to the component's first store.

---

## Combining Stores and Signals

Use stores as the source of truth, signals for local interactivity:

```go
type ThemeSwitch struct {
    composition.Component

    Theme    types.String
    Settings *types.Store
}

func (t *ThemeSwitch) OnMount() {
    state.Effect(func() func() {
        t.Settings.Set("theme", t.Theme.Get())
        return nil
    })
}
```

The store persists the theme; the signal drives local reactivity. This pattern balances global consistency with component responsiveness.

---

## API Summary

### Signals (`github.com/rfwlab/rfw/v2/types`)

```go
count := types.NewInt(0)
name  := types.NewString("Ada")
flag  := types.NewBool(false)
val   := types.NewFloat(1.5)
any   := types.NewAny(nil)
```

Common methods: `Get()`, `Set()`, `Read() any`, `OnChange()`, `Channel()`. All nil-safe.

### Host Signal Types

```go
// For SSC host-synced values
type HInt    // *Signal[int] + host binding
type HString // *Signal[string] + host binding
type HBool   // *Signal[bool] + host binding
type HFloat  // *Signal[float64] + host binding
```

### Stores (`github.com/rfwlab/rfw/v2/state`)

```go
s := state.NewStore("name", state.WithModule("app"), state.WithPersistence())

s.Set("key", value)
s.Get("key")

s.OnChange("key", func(v any) { /* ... */ })

undo, redo := s.Undo, s.Redo
state.NewStore("name", state.WithHistory(50))
```

### History (`github.com/rfwlab/rfw/v2/types`)

```go
hist := types.NewHistory(50)  // max 50 snapshots
hist.Bind(myStore)
hist.Snapshot()
hist.Undo()
hist.Redo()
```

---

## Related

- [Quick Start](/docs/getting-started/quick-start)
- [SSC](/docs/guide/ssc)
- [State API](../api/state)