# Global Store vs Local Signals

State in **rfw** can be managed with two complementary tools: **global stores** and **local signals**. Stores act as centralized containers for shared, persistent data. Signals are lightweight reactive variables confined to a component. This guide explains the differences and when to use each.

---

## Global Store

A store holds data shared across your application. Keys are identified by module and name, and components automatically subscribe to the values they use.

### Example

```go
package main

import "github.com/rfwlab/rfw/v1/state"

var profile = state.NewStore("profile", state.WithModule("user"))

func init() {
    profile.Set("first", "Ada")
    profile.Set("last", "Lovelace")
}
```

In templates, `@user/profile.first` binds directly to this store. Updates propagate automatically. Stores also support computed values, watchers, persistence, and undo/redo.

---

## Local Signals

A signal is a reactive variable scoped to the component that creates it. Perfect for transient or UI-only state.

### Example

```go
count := state.NewSignal(0)

stop := state.Effect(func() func() {
    fmt.Println("count:", count.Get())
    return nil
})
defer stop()

count.Set(1)
count.Set(2)
```

Every `Set` re-triggers only the effects that read the signal, avoiding unnecessary recalculations.

---

## Choosing Between Them

| Scenario                              | Use Store  | Use Signal |
| ------------------------------------- | ---------- | ---------- |
| Data shared across many components    | ✅          | ❌          |
| Temporary state in a single component | ❌          | ✅          |
| Persistence in `localStorage`         | ✅          | ❌          |
| Fine-grained reactive updates         | ⚠️ depends | ✅          |

* **Stores** simplify synchronization of complex data across the app.
* **Signals** shine for small, isolated pieces of state.

---

## Combining Stores and Signals

Often you’ll use both: stores as the source of truth, signals for local interactivity.

```go
settings := state.NewStore("settings")
theme := state.NewSignal("light")

state.Effect(func() func() {
    settings.Set("theme", theme.Get())
    return nil
})
```

Here the store persists the theme, while the signal drives local reactivity. This pattern balances global consistency with component responsiveness.
