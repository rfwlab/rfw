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

## Identifiers, variables and commands

RTML distinguishes three categories that should never overlap:

- **Identifiers (`id` attributes)** mark elements so that CSS or
  JavaScript can target them. They are part of plain HTML and RTML never
  substitutes or evaluates their values.
- **Variables (`{expr}` placeholders)** inject dynamic data. They render
  the value of the expression as text and cannot perform logic or alter
  attributes.
- **Commands (`@command` directives)** invoke RTML instructions. They
  add behaviour like loops, conditions or event handlers and may emit
  markup, but they do not expose data directly.

Each construct has a dedicated role and must not take the others' job:

```rtml
<div id="user-card">
  <h2>{user.Name}</h2>
  <button @click:save>Save</button>
</div>
```

- `id="user-card"` is static and only identifies the element.
- `{user.Name}` outputs data but cannot trigger logic.
- `@click:save` registers an event handler but does not display text or
  provide an identifier.

Keeping these responsibilities separate ensures predictable templates
and avoids accidental coupling between style, data and behaviour.

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

Lists can contain component instances. When the loop variable resolves to a
component, using `@prop:var` will render it:

```rtml
@for:item in items
  @prop:item
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

Props and slots may also carry components created in Go. Rendering uses
`@prop:name` or slot placeholders. These components share the parent's store
and, when lists change, **Selective DOM Patching** uses `data-key`
attributes to reuse or reorder items. Without a key the element is
recreated.

## Stores

Access global state using the `@store` directive. This command binds an
element to a value stored in the global state manager and optionally
writes changes back.

The full syntax is:

```
@store:MODULE.STORE.KEY[:w]
```

- `MODULE` selects the module namespace (commonly `app` or `default`).
- `STORE` chooses the store within that module.
- `KEY` is the field to read.
- `:w` is optional and enables two‑way bindings for form controls,
  writing user input back to the store when it changes.

Example:

```rtml
<p>Shared: @store:app.default.sharedState</p>
<input value="@store:app.default.sharedState:w">
```

Outside of form elements the `:w` suffix has no effect and the value is
read‑only.

---

These building blocks cover most RTML syntax. Combine them with Go
logic to create dynamic components.
RTML templates bind data reactively.

@include:ExampleFrame:{code:"/examples/components/computed_component.go", uri:"/examples/computed"}
