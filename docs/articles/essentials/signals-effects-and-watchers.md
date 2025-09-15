# Signals, Effects & Watchers

Signals provide **fine-grained reactive values** that notify only the computations that read them. They are ideal for local component state without needing global stores.

For a comparison with stores, see [Stores vs Signals](../guide/store-vs-signals). For the full API, see the [Signal reference](../api/signal).

---

## Creating a Signal

Use `state.NewSignal` to create a signal and `Get` / `Set` to work with its value:

```go
count := state.NewSignal(0)
count.Set(1)
current := count.Get()
```

Signals are lightweight, reactive containers. Updating a signal re-runs only the computations or template bindings that depend on it.

---

## Running Effects

`state.Effect` registers a function that re-runs whenever any signal it reads changes. It may return a cleanup function that runs before the next effect and when stopped:

```go
stop := state.Effect(func() func() {
    v := count.Get()
    fmt.Println("count is", v)
    return nil
})

defer stop()
count.Set(2) // triggers the effect
```

Effects automatically track dependencies: only the signals read inside cause re-runs.

---

## Watching Signals and Stores

Watchers let you run callbacks when a signal or store value changes. They are useful for side effects like logging, analytics, or interacting with external APIs.

### Watching a Signal

```go
sig := state.NewSignal(0)
state.WatchSignal(sig, func(v int) {
    log.Println("value updated to", v)
})
```

### Watching a Store

```go
s := state.NewStore("counter")
s.Set("count", 0)

stop := state.Watch(state.Path{"default", "counter", "count"}, func(v any) {
    log.Println("count changed to", v)
})
// Call stop() in OnUnmount to avoid leaks
```

Watchers can be configured with options like `Immediate` or `Deep` to run instantly or track nested structures.

---

## Using Signals in Templates

Templates bind to signals with the `@signal:name` directive. The DOM updates whenever the signal changes:

```rtml
<p>Local count: @signal:count</p>
```

### Writable bindings

Form controls can write back to signals by appending `:w`:

```rtml
<input value="@signal:message:w">
<input type="checkbox" checked="@signal:agree:w">
<textarea>@signal:bio:w</textarea>
```

Typing or toggling updates the signal value, and any other bindings update automatically.

@include\:ExampleFrame:{code:"/examples/components/signal\_bindings\_component.go", uri:"/examples/signal-bindings"}

---

## Conditionals and Loops

Signals can drive conditional blocks and loops:

```rtml
@if:signal:count == "3"
  <p>Three!</p>
@endif

@for:item in signal:items
  <li>{item}</li>
@endfor
```

If `items` is a signal holding a slice or map, changes patch only the affected DOM nodes.

---

## Passing Signals as Props

Provide signals through component props so children can bind to them:

```go
c := core.NewComponent("Example", tpl, map[string]any{"count": count})
```

The template can then use `@signal:count` to stay reactive.

---

## End-to-End Example

The following combines writable bindings, conditionals, and a signal-backed slice:

@include\:ExampleFrame:{code:"/examples/components/signals\_effects\_component.go", uri:"/examples/signals"}

---

## Why Use Signals and Watchers

* **Local state**: no need for global stores.
* **Fine-grained updates**: only re-run what depends on the signal.
* **Declarative templates**: signals bind naturally in RTML.
* **Two-way binding**: `:w` makes form inputs reactive out of the box.
* **Side effects**: watchers let you react to changes without polluting templates.
