# Reactivity Fundamentals

RFW reactivity keeps the DOM synchronized with Go data. Any exported field on a component becomes a reactive source. When its value changes, RFW computes the minimal DOM patches and applies them efficiently.

## Reactive State

Fields exported on the component struct are observed automatically. Updating them triggers a re-render of only the nodes that depend on those fields.

```go
type Counter struct {
  *core.HTMLComponent
  Count int
}

func (c *Counter) Inc() {
  c.Count++ // DOM updates automatically
}
```

Bindings in the template use `{}` or directives to read these fields. RFW tracks which nodes consume which fields so unrelated parts of the DOM aren't touched.

## Stores

For global state, create stores with `state.NewStore`. Components bind to store values using the `@store` directive in templates. Store updates notify all subscribed components.

```go
var counter = state.NewStore("counter", state.Map{"count": 0})

func main() {
  state.RegisterModule("default", state.Module{Stores: state.Stores{counter}})
}
```

```rtml
<p>Shared: @store:default.counter.count</p>
```

## Avoiding Unnecessary Work

Because rendering is tied to data dependencies, expensive computations should be placed in computed properties or watchers rather than inline in templates. This ensures RFW only recomputes when required and keeps interfaces snappy.

Understanding how reactive data flows is essential before exploring higher‑level features like computed properties and watchers.
