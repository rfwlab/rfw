# Conditional Rendering

Many interfaces display elements only when certain conditions are met. RTML provides `@if`, `@else-if`, and `@else` directives to declaratively control what appears in the DOM.

## Basic Usage

Wrap markup with `@if:expression` to render it when the expression is truthy:

```rtml
@if:loggedIn
  <p>Welcome back!</p>
@endif
```

The expression has access to component fields and store bindings. When `loggedIn` becomes false, the paragraph is removed from the DOM.

## Else Blocks

Use `@else` or `@else-if` for alternate branches:

```rtml
@if:status == 'loading'
  <p>Loading...</p>
@else-if:status == 'error'
  <p>Failed to load.</p>
@else
  <p>Ready!</p>
@endif
```

Each branch may contain any valid RTML, including loops or component includes.

Conditions can reference signals as well. The block below renders when the
`count` signal holds the string "3":

```rtml
@if:signal:count == "3"
  <p>Three!</p>
@endif
```

## Combining with Loops

Conditionals can be nested inside `@for` loops or vice versa to render complex structures while keeping templates readable.

Conditional rendering keeps business logic in Go while letting templates describe multiple UI states succinctly.
