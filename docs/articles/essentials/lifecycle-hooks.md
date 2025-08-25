# Lifecycle Hooks

Each component moves through a predictable set of stages. Lifecycle hooks let you run code at key moments without scattering logic across the application.

## Mount

When a component is first rendered, its template is converted to DOM and inserted into the page. The optional `OnMount` hook fires afterward:

```go
func (c *Widget) OnMount() {
  go c.loadData()
}
```

Start timers, fetch remote data, or access refs here. The component is fully present in the DOM.

## Update

Whenever reactive state changes, the component re-renders. After the patch is applied, `OnUpdate` runs:

```go
func (c *Widget) OnUpdate() {
  log.Println("widget updated")
}
```

Use this hook for side effects that must occur after every render.

## Unmount

Before a component is removed, `OnUnmount` executes. Clean up any external resources to avoid leaks:

```go
func (c *Widget) OnUnmount() {
  close(c.done)
}
```

Hooks can be registered by implementing the methods directly or by calling `SetOnMount`, `SetOnUpdate`, and `SetOnUnmount` on the underlying `HTMLComponent`. Lifecycle hooks provide structured entry points for augmenting component behavior.
