# Conditional Rendering

Many UIs need to show or hide content depending on state. RTML supports this directly with `@if:`, `@else-if:`, `@else`, and `@endif` directives, no manual DOM updates needed.

---

## Basic Usage

Wrap markup with `@if:expression` to render only when the expression is truthy:

```rtml
@if:loggedIn
  <p>Welcome back!</p>
@endif
```

When `loggedIn` is falsy, the `<p>` is removed from the DOM. Expressions can reference component fields, props, signals, or store values.

---

## Else Blocks

Use `@else` and `@else-if` to add alternate branches:

```rtml
@if:status == "loading"
  <p>Loading...</p>
@else-if:status == "error"
  <p>Failed to load.</p>
@else
  <p>Ready!</p>
@endif
```

Each branch can contain any valid RTML, including loops and nested components.

---

## Signals in Conditions

Use `@signal:Name` inside conditions to check reactive values:

```rtml
@if:signal:Count == "3"
  <p>Three!</p>
@endif
```

The block updates automatically whenever the `Count` signal changes. The comparison uses string coercion, signal values are compared as strings in `@if:` expressions.

---

## @expr: in Conditions

For computed or numeric comparisons, use `@expr:` expressions:

```rtml
@if:@expr:Count.Get > 5
  <p>More than five!</p>
@endif

@if:@expr:Count.Get == 0
  <p>Zero!</p>
@else
  <p>Non-zero</p>
@endif
```

`@expr:` supports arithmetic (`+`, `-`, `*`, `/`), comparisons (`==`, `!=`, `<`, `>`, `<=`, `>=`), and logical operators (`&&`, `||`, `!`). Field access like `.Get` resolves the signal's current value.

---

## Ternary in @expr:

Use inline conditionals with `then ... else` inside `@expr:`:

```rtml
<p>Status: @expr:Active.Get then "Online" else "Offline"</p>
<p>Size: @expr:Count.Get > 10 then "large" else "small"</p>
```

Legacy `? :` syntax also works:

```rtml
<p>Status: @expr:Active.Get ? "Online" : "Offline"</p>
```

Prefer `then ... else` for readability in templates.

---

## Negation

Use `!` to invert a condition:

```rtml
@if:!loggedIn
  <p>Please sign in.</p>
@endif
```

---

## Combining with Loops

Conditionals and loops can be nested in either order:

```rtml
@for:item in signal:Items
  @if:item.Done
    <li class="done">{item.Text}</li>
  @else
    <li>{item.Text}</li>
  @endif
@endfor
```

---

## Summary

| Directive        | Purpose                          |
| ---------------- | -------------------------------- |
| `@if:expr`       | Render block when expr is truthy |
| `@else-if:expr`  | Alternate branch                  |
| `@else`          | Fallback branch                   |
| `@endif`         | Close the conditional block       |

Signal-based conditions with `@signal:` update reactively. Computed conditions with `@expr:` support full arithmetic and comparison expressions.