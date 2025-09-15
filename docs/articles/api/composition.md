# composition

Utilities for composing DOM nodes and binding elements.

| Function | Description |
| --- | --- |
| `Wrap(c *core.HTMLComponent) *Component` | Wrap an `HTMLComponent` to access composition helpers. |
| `FromProp[T any](c *Component, key string, def T) *state.Signal[T]` | Create a signal from a component prop. |
| `Group(nodes ...Node) *Elements` | Collect multiple nodes. |
| `Bind(selector string, fn func(El))` | Apply a function to elements matching a selector. |
| `BindEl(el dom.Element, fn func(El))` | Apply a function to a specific element. |
| `For(selector string, fn func() Node)` | Repeat node creation for matching elements. |
| `Div()` | Build a `<div>` element node. |
| `A()` | Build an `<a>` element node. |
| `Span()` | Build a `<span>` element node. |
| `Button()` | Build a `<button>` element node. |
| `H(level int)` | Build a heading element. |

