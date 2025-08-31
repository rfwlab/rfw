# Global Store vs Local Signals

State management in rfw can be handled through two complementary mechanisms: the **global store** and **local signals**. The store provides a centralized, persistent container, ideal for data shared across multiple components, while signals offer fine-grained reactivity limited to the context in which they are created. This guide highlights the main differences and shows when to choose one over the other.

## Global Store

The store is designed to maintain data shared throughout the application. Each key is identified by a module and a name, allowing components to subscribe only to the information they need.

### Example

```go
package main

import (
    "github.com/rfwlab/rfw/v1/state"
)

var profile = state.NewStore("profile", state.WithModule("user"))

func init() {
    profile.Set("first", "Ada")
    profile.Set("last", "Lovelace")
}
```

Components can read values with `@user/profile.first` and will be automatically updated when they change. Stores also support computed values and watchers to react to updates.

## Local Signals

Signals provide a reactive variable confined to the component that creates it. They are perfect for transient state, such as input fields or UI flags.

### Example

```go
package main

import (
    "fmt"
    "github.com/rfwlab/rfw/v1/state"
)

func main() {
    count := state.NewSignal(0)

    stop := state.Effect(func() func() {
        fmt.Println("count:", count.Get())
        return nil
    })
    defer stop()

    count.Set(1)
    count.Set(2)
}
```

Each call to `Set` notifies only the functions that have read the signal, avoiding unnecessary recalculations.

## When to Use Each

| Scenario                                 | Global Store       | Local Signals |
| ---------------------------------------- | ------------------ | ------------- |
| Data shared across many parts of the app | ✅                  | ❌             |
| Temporary state of a single component    | ❌                  | ✅             |
| Persistence in `localStorage`            | ✅                  | ❌             |
| Fine-grained updates                     | ⚠️ depends on keys | ✅             |

The store simplifies synchronization of complex data, while signals shine in handling small, isolated pieces of state.

## Combining Store and Signals

It’s common to use both. A store can provide main data, while a component creates derived signals for local interaction.

```go
package main

import "github.com/rfwlab/rfw/v1/state"

func example() {
    settings := state.NewStore("settings")
    theme := state.NewSignal("light")

    state.Effect(func() func() {
        current := theme.Get()
        settings.Set("theme", current)
        return nil
    })
}
```

This way the store remains the single source of truth, but the signal enables local reactive flow.
