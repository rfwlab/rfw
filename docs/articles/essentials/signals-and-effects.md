# Signals & Effects

Signals provide fine-grained reactive values that notify only the computations
that read them. They are ideal for local component state without relying on
store keys.

## Creating a Signal

Use `state.NewSignal` to create a signal and `Get`/`Set` to work with its value:

```go
count := state.NewSignal(0)
count.Set(1)
current := count.Get()
```

## Running Effects

`state.Effect` registers a function that re-runs whenever any signal read inside
changes. The function may return a cleanup callback which executes before the
next run and when the effect is stopped.

```go
stop := state.Effect(func() func() {
    v := count.Get()
    fmt.Println("count is", v)
    return nil
})

defer stop()
count.Set(2) // triggers the effect
```

Effects automatically track which signals they access, so unrelated signals do
not cause extra work.

## Using Signals in Templates

Components can bind template expressions to signals using the `@signal:name`
directive. The DOM updates whenever the signal's value changes.

```rtml
<p>Local count: @signal:count</p>
```

Provide the signal through component props:

```go
c := core.NewComponent("Example", tpl, map[string]any{"count": count})
```

The template above will render the current value and update automatically when
the `count` signal changes.

@include:ExampleFrame:{code:"/examples/components/signals_effects_component.go", uri:"/examples/signals"}
