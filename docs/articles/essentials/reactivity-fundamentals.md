# Reactivity Fundamentals

Reactivity is the **core mechanism** that makes rfw applications dynamic. It ensures that the DOM always reflects the current Go state—no manual DOM updates, no virtual DOM overhead. Understanding this system is essential before moving on to advanced features.

---

## How It Works

Every reactive value in rfw is **tracked**. When a value changes, rfw determines exactly which DOM nodes depend on it and applies **minimal patches**. Only the parts of the interface that need updating are touched.

This fine-grained approach keeps rendering efficient even in large applications.

---

## Reactive Fields in Components

Any **exported field** on a component struct is automatically reactive:

```go
type Counter struct {
  *core.HTMLComponent
  Count int
}

func (c *Counter) Inc() {
  c.Count++ // DOM updates automatically
}
```

RTML templates bind to these fields using `{}` placeholders or directives. When `Count` changes, rfw patches only the text node that renders it—other parts of the DOM remain untouched.

Key points:

* Exported fields are reactive by default.
* Updates are local: only dependent nodes are re-rendered.
* No explicit setters or signals are required for basic cases.

---

## Signals

For more control, rfw provides **signals**: standalone reactive values created with `state.NewSignal`. Signals are ideal for local, fine-grained state not tied to a struct field.

```go
count := state.NewSignal(0)
count.Set(count.Get() + 1)
```

Signals integrate seamlessly with templates via props or composition. They are lighter than global stores and perfect for per-component state.

In templates, signals bind with `@signal:name`. For form elements, append `:w` to make the binding **writable**:

```rtml
<p>@signal:message</p>
<input value="@signal:message:w">
<input type="checkbox" checked="@signal:agree:w">
<textarea>@signal:bio:w</textarea>
```

Typing into these controls updates the signal, and any `@signal` bindings elsewhere update automatically.

---

## Global Stores

When state must be shared across components, use a **store**. Stores are created with `state.NewStore` and hold reactive key/value pairs.

```go
var counter = state.NewStore("counter")
counter.Set("count", 0)
```

Templates bind to stores using the `@store` directive:

```rtml
<p>Shared: @store:default.counter.count</p>
<input value="@store:default.card.address.zip:w">
```

Stores:

* Provide a single source of truth for global data.
* Notify all subscribed components when values change.
* Can be persisted to `localStorage` or enriched with computed values and watchers.

Appending `:w` enables **two-way binding** on form elements so user input writes back to the store.

---

## Dataflow and Dependencies

rfw builds a **dependency graph** between state and DOM nodes:

* When a field, signal, or store value is read in a template, rfw records the dependency.
* On update, only the affected nodes are recomputed.
* Unrelated parts of the UI remain stable.

This ensures performance scales even as applications grow.

---

## Avoiding Unnecessary Work

Reactivity is powerful, but misuse can lead to inefficiency:

* Avoid placing expensive calculations inline in templates.
* Use **computed properties** to cache derived values until their dependencies change.
* Use **watchers** for side effects, logging, or syncing external systems.

This pattern guarantees that costly work only runs when necessary.

---

## Why This Matters

* **Predictability**: UI always reflects Go state.
* **Efficiency**: DOM updates are minimal, avoiding full re-renders.
* **Scalability**: Works for both small widgets and large apps.
* **Simplicity**: No manual DOM manipulation; just update Go values.

---

## Summary

* Exported struct fields are reactive by default.
* Signals provide lightweight local reactivity, with `:w` enabling write-back in forms.
* Stores provide shared global state, with optional writable bindings for inputs.
* rfw tracks dependencies and patches only what changed.
* Computed properties and watchers prevent unnecessary work.

Mastering these fundamentals is the key to writing performant, maintainable rfw applications.
