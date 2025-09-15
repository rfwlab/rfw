# Computed Properties

Computed properties let you derive new values from existing reactive state. They automatically update when their dependencies change and cache the result until then. This ensures you avoid redundant recalculations.

---

## Defining a Computed Value

Use `state.NewComputed` to declare a computation. It runs once, and reruns only when any of its dependencies change:

```go
var double = state.NewComputed(func() any {
  return store.Get("count").(int) * 2
})
```

In RTML you can bind it like any other field:

```rtml
<p>Doubled: {double.Value()}</p>
```

The call to `Value()` reads the cached result. When `count` changes, the cache is invalidated and the function recomputes on next access.

---

## Using Computed in Components

You can define computed values inside components and expose them as fields:

```go
type Stats struct {
  *core.HTMLComponent
  Count int
  Even  state.Computed
}

func NewStats() *Stats {
  s := &Stats{HTMLComponent: core.NewComponent("Stats", tpl, nil)}
  s.Even = state.NewComputed(func() any { return s.Count%2 == 0 })
  s.SetComponent(s)
  s.Init(nil)
  return s
}
```

```rtml
<p>Is even? {Even.Value()}</p>
```

---

## Why Use Computed

* Keep templates simple by moving logic into Go functions.
* Ensure derived values are always in sync with their dependencies.
* Avoid repeating expensive calculations on every render.

Computed properties are especially useful for data transformations, formatting, or combining multiple signals/stores into one reactive output.
