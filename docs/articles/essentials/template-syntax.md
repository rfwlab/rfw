# Template Syntax

RTML is RFW's declarative, HTML-like language for describing user interfaces. It extends standard markup with directives that connect the DOM to Go data and events. Templates are compiled to Go code so no parser runs in the browser. Browser integration currently relies on plain JavaScript; TypeScript builds are not supported yet, though a future release will expose a global `rfw` object to interact with RFW APIs directly.

## Variables, Commands and Constructors

RTML distinguishes three constructs:

- **Variables**: placeholders wrapped in `{}` that insert reactive values.
- **Commands**: directives starting with `@` that drive logic or register behavior.
- **Constructors**: square‑bracket annotations inside a start tag. `[name]` on an element adds a reference, while `[key {expr}]` injects a list key for efficient diffing.

Each construct has its own responsibility and they should not be intermixed. Example:

```rtml
<div [userCard]>
  <h2>{user.Name}</h2>
  <button @on:click:save>Save</button>
</div>
```

`[userCard]` registers the `<div>` for later lookup with `GetRef`, `{user.Name}` prints data, and `@on:click:save` attaches an event handler.

## Text Interpolation

Place `{expression}` inside markup to insert reactive data:

```rtml
<p>Count: {count}</p>
```

Changing the `count` field on the component automatically updates the rendered text.

### Calling Functions

Expressions can invoke methods that are exposed on the component:

```rtml
<time>{formatDate(createdAt)}</time>
```

Functions are evaluated every render so they should be side-effect free.

## Attribute and Store Bindings

Attributes may also contain expressions:

```rtml
<img src="{user.Avatar}" alt="{user.Name}">
```

For global state, use the `@store` command to bind a value to an attribute. The optional `:w` suffix enables two‑way binding for form controls. A `:r` suffix is not recognized; omit the suffix for read‑only bindings:

```rtml
<input value="@store:default.counter.count:w">
<input type="checkbox" checked="@store:default.counter.enabled:w">
<textarea>@store:default.counter.notes:w</textarea>
```

Updating the `count` key in the `counter` store reflects in the input, and editing the input writes back to the store.

Local signals may be bound similarly using `@signal:name` and `@signal:name:w` for two‑way bindings:

```rtml
<p>@signal:message</p>
<input value="@signal:message:w">
<input type="checkbox" checked="@signal:agree:w">
<textarea>@signal:bio:w</textarea>
```

### Boolean Attributes

When an expression resolves to a boolean, the attribute is included only if the value is truthy:

```rtml
<button disabled="{isDisabled}">Save</button>
```

If `isDisabled` is `false`, the `disabled` attribute is removed from the element.

### Keyed Bindings

Lists often need stable identity for efficient updates. Use the `[key {expr}]` constructor inside the looped element so RFW can patch DOM nodes selectively:

```rtml
@for:item in items
  <li [key {item.ID}]>{item.Text}</li>
@endfor
```

Without a key, list items are recreated when their order changes.

## Event Handling

Events are bound with the `@on:` prefix followed by the event name and handler:

```rtml
<button @on:click:increment>Increment</button>
```

### Event Modifiers

Modifiers may be appended after the event name to adjust behavior:

| Modifier | Description |
|----------|-------------|
| `stop` | Calls `event.stopPropagation()` to prevent bubbling. |
| `prevent` | Calls `event.preventDefault()` to stop the browser's default action. |
| `once` | Removes the listener after the first invocation. |

Example:

```rtml
<form @on:submit.prevent.stop:onSubmit>
<button @on:click.once:launch>Launch once</button>
```

Event handlers are registered with `dom.RegisterHandlerFunc` on the Go side.

## Conditionals

`@if`, `@else-if`, and `@else` conditionally render blocks:

```rtml
@if:count > 0
  <p>Positive</p>
@else
  <p>Zero or negative</p>
@endif
```

Conditional blocks may contain any valid RTML, including loops or components.

## Lists

Iterate collections with `@for`:

```rtml
@for:item in store:default.todos.items
  <li>{item.text}</li>
@endfor
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

When a signal's `Read` method returns a slice or map, `@foreach` can iterate it:

```rtml
@foreach:items as it
  <li>@it</li>
@endforeach
```

## Components, Props and Slots

Bring in child components with `@include:Name`, optionally passing props, and expose content placeholders with `@slot`:

```rtml
@include:Card:{title:"Hello"}

@slot:Card.body
  <p>Content from parent</p>
@endslot
```

Properties on the included component are accessed with `@prop`:

```rtml
<card>
  <h2>@prop:title</h2>
</card>
```

Slots can receive markup or even other components, allowing parents to control portions of a child component's template.

## Dynamic Components

Use `rt-is` to switch components at runtime based on a store value or prop:

```rtml
<div rt-is="{current}"></div>
```

The element is replaced with the component whose name matches `current`.

## Using Expressions

RTML evaluates JavaScript-like expressions within `{}` and in directive values:

```rtml
<p>{user.First + " " + user.Last}</p>
<div class="btn-{variant}"></div>
```

Expressions must resolve to a value; statements like `if` or `var` are invalid. Only a restricted set of global functions is available, such as `Math` and `Date`.

## Directives Summary

RTML ships with a small set of built-in directives:

- `@on:event:handler` – attach event handlers.
- `@if`, `@else-if`, `@else` – conditional rendering.
- `@for` – iterate over collections or ranges.
- `@include:Component` – render another component.
- `@slot:name` / `@endslot` – declare named slots inside a component.
- `@prop:name` – read a property passed to a component.
- `@store:module.store.key[:w]` – bind to global state.
- `@signal:name[:w]` – bind to a local signal.
- `rt-is` – render a component dynamically.

Commands may accept parameters or modifiers; see individual sections above for details. Plugins can also contribute `{plugin:NAME.var}` variables, `@plugin:NAME.cmd` commands and `[plugin:NAME.ref]` constructors. See the [Plugin API](../api/plugins.md#rtml-directives) and the [Plugin Directives](../guide/plugin-directives.md) guide for details and examples.

## Stores

Access global state using the `@store` directive. This command binds an element to a value stored in the global state manager and optionally writes changes back.

```
@store:MODULE.STORE.KEY[:w]
```

- `MODULE` selects the module namespace (commonly `app` or `default`).
- `STORE` chooses the store within that module.
- `KEY` is the field to read.
- `:w` is optional and enables two‑way bindings for form controls.
- `:r` is not supported; omit the suffix for read‑only bindings.

Example:

```rtml
<p>Shared: @store:app.default.sharedState</p>
<input value="@store:app.default.sharedState:w">
```

Outside of form elements the `:w` suffix has no effect and the value is read‑only. There is no `:r` suffix.

## Security Notes

Interpolated data is escaped by default to prevent XSS vulnerabilities. Only include trusted content and avoid attempting to render arbitrary HTML. Prefer components and slots for rich markup.

---

These building blocks cover most of RTML's template syntax. Combine them with Go logic to create dynamic components. Templates compile to Go, which registers dependencies and subscribes to updates, ensuring the DOM stays in sync with your data.

