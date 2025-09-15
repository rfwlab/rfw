# Composition API

The **composition package** is the foundation of the Composition API in rfw. It wraps a base `*core.HTMLComponent` and exposes helpers for props, events, DOM bindings, groups of elements, and local state.

---

## Why wrap components?

By default, `core.NewComponent` only returns a raw HTML component. To make it usable with the Composition API, you must **wrap** it with `composition.Wrap`. The wrapper:

* Adds helpers like `Prop`, `FromProp`, `On`, `GetRef`, `Group`, and `Store`.
* Ensures the component can expose reactive values and register event handlers.

It is not automatic because some developers still prefer struct components. Wrapping makes the choice explicit: if you want Composition-style features, you call `Wrap`. If not, you stick with the raw component or a struct.

---

## Wrapping Components

```go
hc := core.NewComponent("Name", nil, nil)
cmp := composition.Wrap(hc)
```

The wrapper does not change rendering logic—it only gives you convenient helpers. Call `Unwrap()` to access the original `*core.HTMLComponent` if needed.

---

## Props

Expose reactive state to RTML with `Prop` and consume it with `FromProp`:

```go
count := state.NewSignal(0)
cmp.Prop("count", count)

// get an existing signal or wrap a plain prop value
other := composition.FromProp[int](cmp, "other", 1)
other.Set(2)
```

* `Prop(key, signal)`: exports a reactive signal under a key.
* `FromProp[T](key, default)`: retrieves a signal from props, or wraps a plain value if found.

Props are immutable for the child—update via parent events or stores.

---

## Event Handlers

Register handlers directly on the component:

```go
cmp.On("save", func() { /* handle save */ })
```

Templates can call these handlers with `@on:click:save`. Internally this forwards to the `dom` package.

If you are not using `composition.Wrap`, you can attach listeners manually via `dom.ByID(...).On(...)`.

---

## DOM Bindings

Sometimes you need direct access to nodes. RTML lets you annotate elements with a constructor:

```rtml
<root>
  <div [list]></div>
</root>
```

Fetch and update it in Go:

```go
cmp.SetOnMount(func(*core.HTMLComponent) {
    listEl := cmp.GetRef("list")
    listEl.SetHTML("")
    listEl.AppendChild(composition.Div().Text("Item").Element())
})
```

* `GetRef(name)`: fetches a node marked with `[name]`.
* `Bind` / `For`: query nodes by selector when you don’t have a ref.

---

## Element Groups

When you create multiple nodes with the builders (`Div()`, `Span()`, etc.), you can group them:

```go
cards := composition.Group(
    composition.Div().Text("A"),
    composition.Div().Text("B"),
)
cards.AddClass("card").SetAttr("data-role", "item")
```

Groups let you:

* Apply bulk operations (`AddClass`, `SetStyle`, etc.).
* Merge multiple groups.
* Iterate over elements with `ForEach`.

---

## Stores and History

Create a store scoped to a component:

```go
s := cmp.Store("count", state.WithHistory(5))
s.Set("v", 1)
s.Set("v", 2)

cmp.History(s, cmp.ID+":undo", cmp.ID+":redo")
```

* `Store(name, opts...)`: creates a component-scoped store.
* `History`: registers undo/redo handlers tied to that store.

This lets you build isolated logic with its own undo/redo flow.

---

## Summary

Wrapping is explicit because rfw supports two styles:

* **Struct components**: embed `*core.HTMLComponent` in a struct.
* **Composition API**: wrap with `composition.Wrap` to get reactive helpers.

Both work, but if you want fine-grained reactivity and helper methods, use the Compositio
