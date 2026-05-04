# Computed Properties

v2 replaces most standalone computed properties with `@expr:` in RTML templates. For complex logic, use Go methods.

---

## @expr: in Templates

Inline computed expressions that re-evaluate when their signal dependencies change:

```rtml
<p>Doubled: @expr:Count.Get * 2</p>
<p>Positive: @expr:Count.Get > 0</p>
<p>Greeting: @expr:Name.Get + "!"</p>
<p>Label: @expr:Active.Get ? "on" : "off"</p>
```

Any `Get()` call inside `@expr:` creates a dependency. When the signal updates, only the affected expression re-renders.

---

## Complex Logic in Go Methods

When `@expr:` isn't enough, define a method on the struct:

```go
type Stats struct {
    *core.HTMLComponent
    Scores *composition.String `rfw:"signal"`
}

func (s *Stats) Average() string {
    // complex logic here
    return result
}
```

In RTML:

```rtml
<p>Average: {Average}</p>
```

The `{Method}` syntax calls the method and injects the result.

---

## When to Use What

| Approach | Use for |
| --- | --- |
| `@expr:` | Simple arithmetic, comparisons, string concatenation |
| Go method + `{Method}` | Multi-step logic, loops, error handling |
| `state.Effect` | Side effects that react to signal changes |

---

## Effects for Derived State

When you need to react to changes with side effects (logging, syncing, DOM manipulation), use `state.Effect`:

```go
state.Effect(func() func() {
    val := count.Get()
    fmt.Println("count changed to", val)
    return nil
})
```

Effects re-run when any signal read inside them changes. Return a cleanup function if needed.