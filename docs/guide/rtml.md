# RTML Templates

RTML is a lightweight template language that looks like HTML but adds a
few conveniences for reactive Go applications.

- **Bindings** are written with `{expr}` and update when the referenced
  state changes.
- **Directives** such as `@click:handler` or `@if:condition` attach event handlers and
  conditional rendering logic.
- The template is compiled to Go code so no runtime parser runs in the
  browser.

Because RTML compiles to Go, complex components still benefit from the
compiler's type checking and tooling support.

## Basic structure

```rtml
<root>
  <h1>Hello {name}</h1>
  <button @click:increment>Count: {count}</button>
</root>
```

## Data bindings

Place `{expr}` inside markup to insert reactive values. The DOM updates
when the expression changes.

```rtml
<p>{user.Name}</p>
```

## Directives

Attributes prefixed with special directives add reactive behaviour.

- `@EVENT:handler` wires event handlers:

  ```rtml
  <button @click:handle>Click</button>
  ```

  The handler name must match a function registered with `dom.RegisterHandlerFunc`.

- `@if:COND`, `@else-if:COND`, `@else` render blocks conditionally:

  ```rtml
  @if:loggedIn
    <p>Welcome back!</p>
  @else
    <p>Please sign in.</p>
  @endif
  ```

## Loops

Iterate over collections or ranges with `@for`.

```rtml
<ul>
  @for:item in items
    <li>{item}</li>
  @endfor
</ul>
```

`@for` also supports key/value pairs and numeric ranges:

```rtml
@for:key,val in obj
  <p>{key}: {val}</p>
@endfor

@for:i in 0..n
  <span>{i}</span>
@endfor
```

## Components, props and slots

Include other templates with `@include` and read properties using
`@prop`.

```rtml
@include:card
<card>
  <h2>@prop:title</h2>
</card>
```

Slots allow passing markup to child components:

```rtml
@slot:avatar
  <img src="{user.Avatar}">
@endslot
@include:userCard
```

## Stores

Access global state using the `@store` directive:

```rtml
<p>Count: @store:app.counter</p>
<input value="@store:app.counter:w">
```

The `:w` suffix writes back to the store when the value changes.

---

These building blocks cover most RTML syntax. Combine them with Go
logic to create dynamic components.
