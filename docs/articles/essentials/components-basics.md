# Components Basics

Components are the fundamental unit of UI in **rfw**. Each component pairs a Go implementation with an **RTML** template. Reactive data is bound into the template; when that data changes, rfw patches only the affected DOM nodes—no virtual DOM required.

---

## Two authoring styles

rfw supports two ways to author components. You can mix them in the same app.

### 1) Composition API (recommended)

Use the `composition` helpers to define reactive state and event handlers close to the template.

```go
//go:build js && wasm
package components

import (
    _ "embed"
    core "github.com/rfwlab/rfw/v1/core"
    "github.com/rfwlab/rfw/v1/composition"
    "github.com/rfwlab/rfw/v1/state"
)

//go:embed templates/counter.rtml
var counterTpl []byte

func NewCounter() *core.HTMLComponent {
    cmp := composition.Wrap(core.NewComponent("Counter", counterTpl, nil))

    count := state.NewSignal(0)
    cmp.Prop("count", count)                  // expose to the template
    cmp.On("inc", func() { count.Set(count.Get()+1) })
    cmp.On("dec", func() { count.Set(count.Get()-1) })

    return cmp.HTML()
}
```

```rtml
<root>
  <button @on:click:dec>-</button>
  <span>{count}</span>
  <button @on:click:inc>+</button>
</root>
```

**Key points**

* `state.Signal[T]` drives fine‑grained updates.
* `cmp.Prop(key, signal)` exposes a reactive value to RTML.
* `@on:click:name` binds DOM events to handlers registered via `cmp.On(name, fn)`.

### 2) Struct components

Embed `*core.HTMLComponent` into a Go struct. Exported fields are reactive and can be referenced by name in RTML.

```go
//go:build js && wasm
package components

import (
    _ "embed"
    core "github.com/rfwlab/rfw/v1/core"
)

//go:embed templates/counter.rtml
var counterTpl []byte

type Counter struct {
    *core.HTMLComponent
    Count int
}

func NewCounterStruct() *Counter {
    c := &Counter{HTMLComponent: core.NewComponent("Counter", counterTpl, nil)}
    // If required by your setup, connect the struct instance to the component
    // so exported fields/methods are observable by the runtime.
    // (Exact wiring details depend on your project’s conventions.)
    return c
}

func (c *Counter) Inc() { c.Count++ }
func (c *Counter) Dec() { c.Count-- }
```

```rtml
<root>
  <button @on:click:Dec>-</button>
  <span>{Count}</span>
  <button @on:click:Inc>+</button>
</root>
```

**Key points**

* Exported fields (e.g., `Count`) are observable by the template.
* Exported methods can be called from RTML via `@on:`.

---

## Props (one‑way data flow)

Props pass data **into** a child component. In RTML:

```rtml
@include:ChildCounter:{start: 5}
```

In code (Composition API):

* `Prop(key, signal)` stores a reactive value under a key.
* `FromProp[T](key, default)` retrieves a `Signal[T]` from props; if the prop is a plain value of type `T`, it’s wrapped into a new signal.

Props are immutable from the child’s perspective. To communicate back **up**, emit events (handlers registered by the parent) or use shared stores.

---

## Lifecycle hooks

Use lifecycle hooks to wire effects, timers, or subscriptions. Ensure you clean up on unmount.

```go
cmp.SetOnMount(func(*core.HTMLComponent) {
    stop := state.Effect(func() func() {
        // read signals / stores here
        return nil // return a disposer if you allocate resources
    })
    cmp.SetOnUnmount(func(*core.HTMLComponent) { if stop != nil { stop() } })
})
```

* **Mount**: attach listeners, kick off async work, preload assets.
* **Unmount**: cancel timers, dispose effects, unsubscribe from stores.

---

## Slots & composition

Components can include other components and expose slots. Parents inject markup into named outlets without coupling logic.

```rtml
<!-- Parent -->
@include:Panel:{title: "Dashboard"}
  <div slot="body">…content…</div>
@end
```

Children render slot content where appropriate. (Slot syntax and capabilities follow RTML rules.)

---

## Reactivity model (under the hood)

* RTML binds attributes, text nodes, and event handlers to **signals** or **exported fields**.
* When a bound value changes, rfw patches only the affected nodes.
* No virtual DOM is used; updates are direct and localized.

---

## Choosing a style

* **Composition API**: prefer for new code; explicit, scalable, and easy to unit test.
* **Struct components**: convenient when you like method receivers and exported fields.

Both interoperate seamlessly—pick the one that fits each component.

---

## See also

* [Template Syntax](/docs/essentials/template-syntax)
* [Signals & Effects](/docs/essentials/signals-and-effects)
* [State Management](/docs/essentials/state-management)
* [Composition](/docs/essentials/composition)
