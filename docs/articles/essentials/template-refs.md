# Template Refs

Sometimes a component needs direct access to a DOM element or child component instance. Template refs offer an escape hatch while keeping most logic declarative.

## Creating a Ref

Annotate an element by placing a constructor inside its start tag:

```rtml
<input [nameInput]>
```

In the component, call `GetRef("nameInput")` after the element mounts:

```go
func (c *Form) OnMount() {
  el := c.GetRef("nameInput")
  events.Listen("focus", el)
}
```

The returned `dom.Element` can be used with low-level DOM helpers for scenarios like focusing, measuring, or integrating third‑party libraries. Template refs should be used sparingly; most interactions can be handled with events and reactive state.
They are typically accessed during `OnMount`; see [Lifecycle hooks](../api/core#lifecycle-hooks) for more on component lifecycles.
