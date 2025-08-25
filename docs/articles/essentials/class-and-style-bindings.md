# Class and Style Bindings

RTML offers concise syntax for toggling CSS classes and inline styles based on reactive data. This keeps presentation logic in templates while avoiding manual DOM manipulation.

## Class Bindings

Bind the `class` attribute to an expression that returns a string or map. String concatenation works for simple cases:

```rtml
<div class="btn btn-{variant}">...</div>
```

For conditional classes, return a map where keys are class names and values are booleans:

```rtml
<div class="{ map('active', isActive, 'disabled', isDisabled) }"></div>
```

Only the keys with truthy values are rendered. The `map` helper is available in template expressions.

## Style Bindings

Inline styles accept a similar map syntax. Keys are CSS properties in camelCase and values are strings or numbers:

```rtml
<div style="{ map('backgroundColor', color, 'width', width + 'px') }"></div>
```

RFW updates only the changed properties, leaving others untouched.

## Binding Objects

When a component exposes a struct of styles or classes, you can reference it directly:

```rtml
<div class="{Styles}"></div>
```

The object should implement `fmt.Stringer` or be convertible to a map for fine‑grained control.

Class and style bindings make it easy to react to state changes with visual feedback without resorting to manual DOM APIs.
