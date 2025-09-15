# Lifecycle Hooks

Components in **rfw** pass through well-defined stages. Lifecycle hooks let you run code at these key moments without scattering logic around. They are essential for data fetching, side effects, and cleanup.

---

## Lifecycle Flow

```
Create → Mount → Update (repeat) → Unmount
```

1. **Create**: the component instance is constructed.
2. **Mount**: template is converted to DOM and inserted into the page.
3. **Update**: reactive state changes trigger DOM patches.
4. **Unmount**: the component is removed from the DOM.

---

## OnMount

Runs after the component is inserted into the DOM:

```go
func (c *Widget) OnMount() {
  go c.loadData()
}
```

Use this to:

* Start timers
* Fetch remote data
* Access refs (`GetRef`) or manipulate child nodes

At this point, the component is fully available in the DOM.

---

## OnUpdate

Runs after each reactive update:

```go
func (c *Widget) OnUpdate() {
  log.Println("widget updated")
}
```

Use this for side effects that must occur after every render, such as syncing with external APIs or libraries.

---

## OnUnmount

Runs before the component is removed:

```go
func (c *Widget) OnUnmount() {
  close(c.done)
}
```

Use this to:

* Stop timers
* Cancel goroutines
* Release watchers or subscriptions

Ensures resources are cleaned up before the component disappears.

---

## Registering Hooks

You can provide hooks by either:

* Implementing the methods directly (`OnMount`, `OnUpdate`, `OnUnmount`), or
* Registering callbacks dynamically:

```go
cmp.SetOnMount(func(*core.HTMLComponent) { ... })
cmp.SetOnUpdate(func(*core.HTMLComponent) { ... })
cmp.SetOnUnmount(func(*core.HTMLComponent) { ... })
```

---

## Why Lifecycle Hooks Matter

Lifecycle hooks are the structured entry points for side effects in rfw. They keep logic predictable:

* **Mount** for setup
* **Update** for reactive side effects
* **Unmount** for teardown

See the [API reference](../api/core#lifecycle-hooks) for all available helpers.
