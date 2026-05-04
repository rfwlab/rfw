# Class and Style Bindings

Dynamic CSS classes and inline styles in RTML, driven by signals and expressions.

---

## Class Bindings

### Static + Signal Interpolation

```rtml
<div class="btn btn-{Variant.Get}">...</div>
```

`{Variant.Get}` substitutes the current value of the `Variant` signal.

### Conditional Classes with @expr:

```rtml
<div class="@expr:Active.Get ? 'active' : ''">...</div>
<div class="@expr:Count.Get > 0 ? 'has-items' : 'empty'">...</div>
```

### Map Syntax

Use the built-in `map()` helper for multiple toggles:

```rtml
<div class="{ map('active', isActive, 'disabled', isDisabled) }"></div>
```

Keys are class names; values are booleans. Only truthy values appear in the output.

---

## Style Bindings

Inline styles support signal interpolation and `@expr:`:

### Signal Interpolation

```rtml
<div style="color: {Color.Get}; font-size: {Size.Get}px">...</div>
```

### @expr: Expressions

```rtml
<div style="@expr:Active.Get ? 'background-color: green' : 'background-color: gray'">...</div>
```

### Map Syntax for Styles

```rtml
<div style="{ map('backgroundColor', color, 'width', width + 'px') }"></div>
```

Keys use camelCase CSS property names. Values can be strings or numbers.

---

## Programmatic Styles via Composition

Use element builders for Go-side style and class manipulation:

```go
composition.Div().
    Class("panel").
    Style("border", "1px solid #ccc").
    Styles("padding", "8px", "margin", "4px")
```

For DOM updates after mount, use refs:

```go
func (c *Card) OnMount() {
    box := c.GetRef("box")
    box.SetStyle("border-color", "red")
    box.AddClass("highlighted")
}
```

Or `Bind` for selector-based updates:

```go
composition.Bind(".panel", func(el composition.El) {
    el.Append(composition.Span().Text("updated"))
})
```

---

## Group Operations

`*Elements` methods operate on all elements in a group:

```go
group := composition.Group(btn1, btn2, btn3)
group.AddClass("visible").
    SetStyle("opacity", "1").
    SetText("ready")
```

---

## Summary

- Use `{Signal.Get}` for simple interpolation in class/style strings.
- Use `@expr:` for conditional and computed values.
- Use `map()` for multi-toggle class/style maps.
- Use refs and composition builders for imperative DOM manipulation.