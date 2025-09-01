# Template Refs

Sometimes a component needs direct access to a DOM element or child component instance. Template refs offer an escape hatch while keeping most logic declarative.

## Creating a Ref

Use the `ref` attribute on an element in RTML:

```rtml
<input ref="nameInput">
```

In the component, call `GetRef("nameInput")` after the element mounts:

```go
func (c *Form) OnMount() {
  el := c.GetRef("nameInput")
  events.Listen("focus", el)
}
```

The returned `dom.Element` can be used with low-level DOM helpers for scenarios like focusing, measuring, or integrating third‑party libraries.

## Component Refs

Refs also work on included components. Adding `ref="child"` to an `@include` exposes the child instance via `GetComponentRef`.

```rtml
@include:Modal ref="modal"
```

```go
func (p *Page) Open() {
  modal := p.GetComponentRef("modal").(*Modal)
  modal.Show()
}
```

Template refs should be used sparingly; most interactions can be handled with events and reactive state.
They are typically accessed during `OnMount`; see [Lifecycle hooks](../api/core#lifecycle-hooks) for more on component lifecycles.
