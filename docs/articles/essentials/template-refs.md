# Template Refs

Template refs are a feature of **constructors** in RTML. Constructors (`[name]`) can serve multiple purposes, one of which is to mark an element (or a child component) for later lookup. This gives your Go code direct access to DOM elements or component instances when you need fine-grained control.

---

## Creating a Ref in Templates

Add a constructor inside an element’s start tag:

```rtml
<input [nameInput]>
```

Inside the component, call `GetRef("nameInput")` to retrieve the reference after the element is mounted:

```go
func (c *Form) OnMount() {
  el := c.GetRef("nameInput")
  el.Focus() // dom.Element supports DOM helpers like focus
}
```

The returned value is a `dom.Element`. Use it for scenarios such as focusing, measuring size, or integrating third-party libraries.

---

## Creating Refs in Go

Refs can also be created directly from Go code:

* **With Composition API** – using `cmp.GetRef("name")` after wrapping a component:

```go
cmp := composition.Wrap(core.NewComponent("Form", tpl, nil))
cmp.SetOnMount(func(*core.HTMLComponent) {
  el := cmp.GetRef("nameInput")
  el.Focus()
})
```

* **Without Composition API** – plain `*core.HTMLComponent` also exposes `GetRef`:

```go
c := core.NewComponent("Form", tpl, nil)
c.SetOnMount(func(*core.HTMLComponent) {
  el := c.GetRef("nameInput")
  el.SetStyle("border", "1px solid red")
})
```

Both approaches rely on `[nameInput]` in the template, but refs can always be retrieved programmatically once the component is mounted.

---

## Child Component Refs

Refs also apply to included components. Adding `[child]` on an `@include` makes the child instance accessible:

```rtml
@include:ChildComponent [child]
```

In Go, you can fetch the wrapped component and call its methods:

```go
child := c.GetRef("child").Component()
child.(*ChildComponent).DoSomething()
```

---

## When to Use Refs

Refs are an **escape hatch**. They should be used sparingly:

* ✅ Focus an input when a form mounts
* ✅ Integrate a third-party widget that expects a DOM node
* ✅ Call imperative methods on a child component
* ❌ Don’t use refs for normal data flow—prefer signals, props, or stores

Most interactivity should stay declarative via events and reactive bindings.

---

## Lifecycle Considerations

Refs are only valid after the element or component is mounted. Access them in `OnMount` or later hooks:

```go
func (c *Widget) OnMount() {
  box := c.GetRef("box")
  box.SetStyle("border", "1px solid red")
}
```

When the component unmounts, refs become invalid.

---

## Refs from Go (without template constructors)

You can **use** refs from Go in both styles (Composition API and struct components), but **creating a named ref** purely from Go is **not specified in documentation**. Template refs are established by placing a constructor in RTML. From Go you have these options:

* **Keep direct handles** to nodes you build with the composition builders:

  ```go
  el := composition.Div().Class("panel").Element() // dom.Element
  el.SetStyle("outline", "1px solid #ccc")
  ```

  You don’t need a ref name if you already hold the element handle.

* **Query the DOM** via selectors using `Bind` / `For` on a wrapped component (Composition API):

  ```go
  cmp.Bind(".panel", func(el dom.Element) { el.AddClass("active") })
  ```

* **Use the low-level DOM helpers** (works with Composition API or plain `*core.HTMLComponent`):

  ```go
  el := dom.ByID("email")
  el.Focus()
  ```

* **Assign IDs / data-attributes** at creation time and retrieve them later with DOM queries.

> If a Go-side API to register a *named* ref exists beyond template constructors, it is **not specified in documentation**.

These approaches work in both Composition API and non-Composition components.

---

## Related Links

* [Template Syntax](./template-syntax)
* [Lifecycle Hooks](./lifecycle-hooks)
* [DOM API](../api/dom)
