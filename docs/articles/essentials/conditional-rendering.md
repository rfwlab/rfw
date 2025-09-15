# Conditional Rendering

Many UIs need to show or hide content depending on state. RTML supports this directly with the `@if`, `@else-if`, and `@else` directives, so you can declare conditions without writing manual DOM updates.

---

## Basic Usage

Wrap markup with `@if:expression` to render only when the expression is truthy:

```rtml
@if:loggedIn
  <p>Welcome back!</p>
@endif
```

When `loggedIn` is false, the `<p>` is removed from the DOM. Expressions can reference component fields, props, signals, or store values.

---

## Else Blocks

Use `@else` and `@else-if` to add alternate branches:

```rtml
@if:status == 'loading'
  <p>Loading...</p>
@else-if:status == 'error'
  <p>Failed to load.</p>
@else
  <p>Ready!</p>
@endif
```

Each branch can contain any valid RTML, including loops and nested components.

---

## Signals in Conditions

Conditions can check reactive signals as well:

```rtml
@if:signal:count == "3"
  <p>Three!</p>
@endif
```

The block updates automatically whenever the `count` signal changes.

---

## Combining with Loops

Conditionals and loops can be nested in either order. This makes it easy to render complex structures while keeping templates declarative and readable.

---

## Why It Matters

Conditional rendering keeps your business logic in Go, while templates handle visual states. This separation leads to cleaner code and UI that adapts automatically to reactive state.
