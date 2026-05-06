# Template Refs

Template refs give Go code direct access to DOM elements. Use them as an escape hatch for imperative operations like focus, measurement, or third-party library integration.

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

### With `*t.Ref` field (recommended)

Declare a `*t.Ref` field on your struct. `composition.New` auto-allocates the ref and resolves it from the DOM on mount:

```go
type Form struct {
    composition.Component
    Input *t.Ref
}

func (f *Form) OnMount() {
    // Input is automatically resolved — call methods on the JS value
    f.Input.Get().Call("focus")
}
```

No manual `GetRef` call needed. The ref is populated during `OnMount`.

### Manual GetRef

For ad-hoc access without a struct field:

```go
func (f *Form) OnMount() {
    input := f.HTMLComponent.GetRef("nameInput")
    input.Call("focus")
}
```

The returned value is a `dom.Element`. It supports DOM helpers like `Focus()`, `SetStyle()`, `AddClass()`, `SetAttr()`.

---

## Child Component Refs

Refs also work on `@include`:

```rtml
@include:Modal [modal]
```

```go
func (p *Page) OnMount() {
    modal := p.GetRef("modal")
    // ...
}
```

---

## OnMount Auto-discovery

When using `composition.New`, methods named `OnMount` and `OnUnmount` with signature `func()` are auto-detected and registered as lifecycle hooks. Refs are resolved before your `OnMount` method runs:

```go
type Widget struct {
    composition.Component
    Box *t.Ref
}

func (w *Widget) OnMount() {
    // Box is already resolved from the DOM
    w.Box.Get().Call("scrollIntoView")
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