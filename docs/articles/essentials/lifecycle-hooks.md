# Lifecycle Hooks

Components in **rfw v2** pass through well-defined stages. Lifecycle hooks let you run code at these key moments, data fetching, timers, cleanup, and other side effects.

---

## Lifecycle Flow

```
Create → Mount → Update (repeat) → Unmount
```

1. **Create**: `composition.New(&struct{})` constructs and auto-wires the component.
2. **Mount**: template is converted to DOM and inserted into the page.
3. **Update**: signal changes trigger reactive DOM patches.
4. **Unmount**: the component is removed from the DOM.

---

## OnMount

Runs after the component is inserted into the DOM. Define it as a **no-argument exported method** on your struct:

```go
package main

import (
    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

type Widget struct {
    composition.Component
    Items *t.Int `rfw:"signal"`
}

func (w *Widget) OnMount() {
    w.Items.Set(10)
}
```

`composition.New` auto-discovers `OnMount()` and registers it, no manual wiring needed.

Use `OnMount` to:

* Start timers or intervals
* Fetch remote data
* Access refs or manipulate child nodes

At this point the component is fully available in the DOM.

---

## OnUnmount

Runs before the component is removed from the DOM. Define it the same way, a no-argument exported method:

```go
func (w *Widget) OnUnmount() {
    close(w.done)
}
```

Auto-discovered by `composition.New`. Use `OnUnmount` to:

* Stop timers
* Cancel goroutines
* Release subscriptions or watchers

---

## SetOnMount / SetOnUnmount

For advanced cases where you need to register hooks dynamically (e.g., in a function-based setup rather than a struct), use `SetOnMount` and `SetOnUnmount` on the underlying `*core.HTMLComponent`:

```go
view := composition.New(&Widget{})
view.SetOnMount(func(_ *core.HTMLComponent) {
    log.Println("mounted")
})
view.SetOnUnmount(func(_ *core.HTMLComponent) {
    log.Println("unmounted")
})
```

This is useful when the hook logic doesn't belong on the struct itself, for example, in layout wrappers or middleware-style components.

---

## Complete Example

```go
type Timer struct {
    composition.Component
    Count *t.Int    `rfw:"signal"`
    done  chan struct{}
}

func (t *Timer) OnMount() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for {
            select {
            case <-ticker.C:
                t.Count.Set(t.Count.Get() + 1)
            case <-t.done:
                ticker.Stop()
                return
            }
        }
    }()
}

func (t *Timer) OnUnmount() {
    close(t.done)
}
```

The framework calls `OnMount` after DOM insertion and `OnUnmount` before removal, guaranteeing cleanup.

---

## Summary

| Hook         | When          | How to register                      |
| ------------ | ------------- | ------------------------------------- |
| `OnMount`    | After DOM insert | `func (c *T) OnMount()`, auto-discovered |
| `OnUnmount`  | Before DOM remove | `func (c *T) OnUnmount()`, auto-discovered |
| `SetOnMount` | After DOM insert | `view.SetOnMount(fn)`, manual        |
| `SetOnUnmount` | Before DOM remove | `view.SetOnUnmount(fn)`, manual     |

Prefer struct methods, `composition.New` discovers them automatically. Use `SetOnMount`/`SetOnUnmount` only when you need dynamic registration.