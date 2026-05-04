# Template Refs

Template refs give Go code direct access to DOM elements or child component instances. Use them as an escape hatch for imperative operations like focus, measurement, or third-party library integration.

---

## Creating Refs in RTML

Add a constructor `[name]` inside an element's start tag:

```rtml
<input [nameInput]>
<div [box]></div>
```

Multiple constructors are allowed:

```rtml
<div [container] [measured]>
```

---

## Accessing Refs in Components

### With composition.New

`composition.New` auto-discovers `OnMount` and `OnUnmount` methods. Refs are accessible via `GetRef` on the embedded `*core.HTMLComponent`:

```go
type Form struct {
    *core.HTMLComponent
    Query *composition.String `rfw:"signal"`
}

func (f *Form) OnMount() {
    input := f.GetRef("nameInput")
    input.Focus()
}
```

### With composition.Wrap

```go
cmp := composition.Wrap(core.NewHTMLComponent("Form", tpl, nil))
cmp.SetOnMount(func(_ *core.HTMLComponent) {
    el := cmp.GetRef("nameInput")
    el.Focus()
})
```

The returned value from `GetRef` is a `dom.Element`. It supports DOM helpers like `Focus()`, `SetStyle()`, `AddClass()`, `SetAttr()`.

---

## Child Component Refs

Refs also work on `@include`:

```rtml
@include:Modal [modal]
```

```go
func (p *Page) OnMount() {
    modal := p.GetRef("modal").Component()
    modal.(*Modal).DoSomething()
}
```

---

## OnMount Auto-discovery

When using `composition.New`, methods named `OnMount` and `OnUnmount` with signature `func()` are auto-detected and registered as lifecycle hooks. No manual `SetOnMount` call needed:

```go
type Widget struct {
    *core.HTMLComponent
    Count *composition.Int `rfw:"signal"`
}

func (w *Widget) OnMount() {
    box := w.GetRef("box")
    box.SetStyle("border", "1px solid red")
}

func (w *Widget) OnUnmount() {
    // cleanup
}
```

---

## Refs from Go (No Template Constructor)

When building elements programmatically, hold direct references instead of using refs:

```go
el := composition.Div().Class("panel").Element()
el.SetStyle("outline", "1px solid #ccc")
```

Use `Bind` or `For` for selector-based DOM queries:

```go
composition.Bind(".panel", func(el composition.El) {
    el.Append(composition.Span().Text("added"))
})
```

Use `dom.ByID("email")` for low-level lookups.

---

## When to Use Refs

- Focus an input on mount
- Integrate a third-party library expecting a DOM node
- Call imperative methods on a child component
- Avoid refs for data flow; use signals, props, or stores instead

---

## Lifecycle

Refs are only valid after mount. Access them in `OnMount` or later. Refs become invalid after unmount.