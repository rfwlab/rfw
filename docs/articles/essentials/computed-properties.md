# Computed Properties

Computed properties derive new values from existing reactive state. They update automatically when their dependencies change and cache the result until then, preventing unnecessary work.

## Defining a Computed Value

Use `state.NewComputed` to declare a function that depends on reactive fields or store values. The computation runs once and reruns only when dependencies change.

```go
var double = state.NewComputed(func() any {
  return store.Get("count").(int) * 2
})
```

Bind the computed value in a template just like any other field:

```rtml
<p>Doubled: {double.Value()}</p>
```

The call to `Value()` reads the cached result. When `count` changes, RFW invalidates the cache and recomputes on next access.

## Using Computed in Components

Inside a component, computed values can be stored as fields:

```go
type Stats struct {
  *core.HTMLComponent
  Count int
  Even state.Computed
}

func NewStats() *Stats {
  s := &Stats{HTMLComponent: core.NewHTMLComponent("Stats", tpl, nil)}
  s.Even = state.NewComputed(func() any { return s.Count%2 == 0 })
  s.SetComponent(s)
  s.Init(nil)
  return s
}
```

```rtml
<p>Is even? {Even.Value()}</p>
```

Computed properties help keep templates simple while ensuring expensive calculations run only when necessary.
