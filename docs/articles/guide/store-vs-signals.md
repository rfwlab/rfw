# Global Store vs Local Signals

State in **rfw v2** is managed with two complementary tools: **global stores** and **local signals**. Stores are centralized containers for shared, persistent data. Signals are lightweight reactive variables scoped to a component. This guide covers the differences and when to use each.

---

## Local Signals

A signal is a reactive variable scoped to the component that creates it. In v2, use `types.NewInt`, `types.NewString`, etc. or `state.NewSignal[T]` directly.

### With `composition.New` and tags

```go
type Counter struct {
    composition.Component

    Count *types.Int `rfw:"signal"`
}
```

`rfw:"signal"` auto‑creates the signal (if nil), registers it as a prop, and wires reactivity. No manual `Prop()` call.

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

Every `Set` re‑triggers only the effects that read the signal.

---

## Global Store

A store holds data shared across your application. Keys are namespaced by module.

### With `rfw:"store"` tag

```go
type Profile struct {
    composition.Component

    Settings *types.Store `rfw:"store:settings"`
}
```

The tag creates (or retrieves) a store named `"settings"` scoped to the component's ID.

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

## Tag Reference

| Tag | Field Type | Effect |
|-----|-----------|--------|
| `rfw:"signal"` | `*types.Int`, `*types.String`, `*types.Bool`, `*types.Float`, `*types.Any` | Auto‑create signal, register as prop |
| `rfw:"signal:name"` | same | Same, with explicit prop name |
| `rfw:"store:name"` | `*types.Store` | Create/retrieve scoped store |

---

## Choosing Between Them

| Scenario | Store | Signal |
|----------|-------|--------|
| Data shared across many components | ✅ | ❌ |
| Temporary state in a single component | ❌ | ✅ |
| Persistence in `localStorage` | ✅ (`WithPersistence`) | ❌ |
| Fine‑grained reactive updates | ⚠️ depends | ✅ |
| Undo/redo support | ✅ (`WithHistory`) | ❌ |
| Server‑side `h:` bindings | ✅ | ❌ |

- **Stores** simplify synchronization of complex data across the app
- **Signals** shine for small, isolated pieces of state

---

## Combining Stores and Signals

Use stores as the source of truth, signals for local interactivity:

```go
type ThemeSwitch struct {
    composition.Component

    Theme    *types.String `rfw:"signal"`
    Settings *types.Store  `rfw:"store:settings"`
}

func (t *ThemeSwitch) OnMount() {
    // Persist signal value into store on change
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

Common methods: `Get()`, `Set()`, `Read() any`

### Stores (`github.com/rfwlab/rfw/v2/state`)

```go
s := state.NewStore("name", state.WithModule("app"), state.WithPersistence())

s.Set("key", value)
s.Get("key")

s.OnChange("key", func(v any) { /* ... */ })

undo, redo := s.Undo, s.Redo
state.NewStore("name", state.WithHistory(50))
```

---

## Related

- [Quick Start](/docs/getting-started/quick-start)
- [SSC](/docs/guide/ssc)
- [State API](../api/state)