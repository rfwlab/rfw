# Template Syntax

RTML is RFW's HTML-like template language. It extends standard markup with directives that connect the DOM to Go data and events. This guide introduces the most common features. Browser integration currently relies on plain JavaScript; TypeScript builds are not supported yet. A future release will expose a global `rfw` object to access RFW APIs directly from JavaScript.

## Text Interpolation

Values from component fields or stores can be inserted with curly braces:

```rtml
<p>Count: {count}</p>
```

Changing the `count` field on the component automatically updates the rendered text.

## Attribute and Store Bindings

Use `@store:module.store.key` to read from a store and `:w` to allow two‑way binding:

```rtml
<input value="@store:default.counter.count:w">
```

Updating the `count` key in the `counter` store reflects in the input, and editing the input writes back to the store.

## Event Handling

Events are bound with the `@on:` prefix followed by the event name and handler:

```rtml
<button @on:click:increment>Increment</button>
```

Modifiers such as `.prevent` or `.once` can be appended (`@on:submit.prevent:onSubmit`).

## Conditionals

`@if`, `@else-if`, and `@else` conditionally render blocks:

```rtml
@if:count > 0
  <p>Positive</p>
@else
  <p>Zero or negative</p>
@endif
```

## Lists

Iterate collections with `@for`:

```rtml
@for:item in store:default.todos.items
  <li>{item.text}</li>
@endfor
```

The directive tracks store changes and re-renders only the affected items.

## Includes and Slots

Bring in child components with `@include:Child`, optionally passing props, and expose content placeholders with `@slot`:

```rtml
@include:Card:{title:"Hello"}

@slot:Card.body
  <p>Content from parent</p>
@endslot
```

## Dynamic Components

Use `rt-is` to switch components at runtime based on a store value or prop:

```rtml
<div rt-is="{current}"></div>
```

RTML's directives compile down to Go code that registers dependencies and subscribes to updates, ensuring the DOM stays in sync with your data.
