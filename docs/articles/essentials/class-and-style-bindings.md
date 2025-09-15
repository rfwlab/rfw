# Class and Style Bindings

In **rfw**, templates can bind CSS classes and inline styles directly to reactive data. This provides a clean way to express presentation logic inside RTML without writing manual DOM updates.

---

## Class Bindings

The `class` attribute accepts either a string or a map expression.

### String Binding

Use string concatenation for simple cases:

```rtml
<div class="btn btn-{variant}">...</div>
```

Here `{variant}` is substituted with the current value of a signal or store property.

### Conditional Classes

For dynamic toggles, return a map where keys are class names and values are booleans:

```rtml
<div class="{ map('active', isActive, 'disabled', isDisabled) }"></div>
```

Only the classes with truthy values appear in the final output. The `map` helper is built into RTML expressions and makes conditional styling concise.

---

## Style Bindings

Inline styles follow the same pattern. Keys are CSS properties (camelCase) and values can be strings or numbers:

```rtml
<div style="{ map('backgroundColor', color, 'width', width + 'px') }"></div>
```

When reactive values change, rfw updates only the affected properties, leaving others untouched.

---

## Binding Objects

Components can also expose entire style or class objects:

```rtml
<div class="{Styles}"></div>
```

The bound object should either:

* implement `fmt.Stringer`, producing a space‑separated class list, or
* be convertible to a map for fine‑grained control.

This allows encapsulating style logic in Go structs and reusing them across templates.

---

## Why It Matters

By integrating class and style bindings into RTML, rfw ensures that UI changes remain declarative and reactive. Developers avoid manual DOM manipulation, reducing boilerplate while keeping state and presentation in sync.
