# Template Syntax

RTML is rfw’s declarative, HTML‑like language for authoring UIs. It extends markup with **variables**, **commands**, and **constructors** that connect the DOM to Go data and events. Templates are compiled to Go—no runtime parser in the browser. (Today integration uses plain JavaScript; a future release will expose a global `rfw` object.)

---

## Building Blocks

RTML distinguishes three constructs, each with a specific role:

* **Variables** — `{expression}` placeholders that insert reactive values.
* **Commands** — directives starting with `@` that control logic or behavior.
* **Constructors** — square‑bracket annotations placed in a start tag to attach metadata to that element.

Example:

```rtml
<div [userCard]>
  <h2>{user.Name}</h2>
  <button @on:click:save>Save</button>
</div>
```

Here `[userCard]` marks the element for lookup via `GetRef` (see **Template Refs**), `{user.Name}` interpolates data, and `@on:click:save` attaches an event handler.

> Keep roles separate: variables render values, commands control flow/behavior, constructors annotate elements.

---

## Text Interpolation

Insert reactive values with `{expression}`:

```rtml
<p>Count: {count}</p>
```

Changing `count` (field, signal, or store‑backed prop) patches only the affected text node.

### Calling Functions

Expressions can call methods exposed by the component:

```rtml
<time>{formatDate(createdAt)}</time>
```

Functions run on render; keep them pure (no side effects) and inexpensive.

---

## Attribute Bindings

Attributes accept expressions too:

```rtml
<img src="{user.Avatar}" alt="{user.Name}">
```

### Boolean Attributes

If an expression evaluates to a boolean, the attribute is present only when truthy:

```rtml
<button disabled="{isDisabled}">Save</button>
```

When `isDisabled` is `false`, the `disabled` attribute is removed.

---

## State Bindings (Stores & Signals)

RTML offers dedicated commands to bind **global stores** and **local signals**.

### Store bindings

```
@store:MODULE.STORE.KEY[:w]
```

* `MODULE` — module namespace (often `default` or `app`).
* `STORE` — store name.
* `KEY` — field in the store.
* `:w` — optional; enables two‑way binding **on form controls** (read‑only elsewhere). There is **no `:r`** suffix—omit for read‑only.

Examples:

```rtml
<p>Shared: @store:default.counter.count</p>
<input value="@store:default.counter.count:w">
<input type="checkbox" checked="@store:default.counter.enabled:w">
<textarea>@store:default.counter.notes:w</textarea>
```

Editing a writable control updates the store key; other bindings update automatically.

### Signal bindings

Bind local reactive state with `@signal:name`:

```rtml
<p>@signal:message</p>
<input value="@signal:message:w">
<input type="checkbox" checked="@signal:agree:w">
<textarea>@signal:bio:w</textarea>
```

Use `:w` on form elements to write back to the signal.

> Tip: Inside `{}` expressions you can reference a signal by its prop name, e.g. `{count}`; for text‑only positions, `@signal:count` is equivalent. Prefer the `{}` form for consistency across bindings.

---

## Conditionals

Render blocks conditionally with `@if`, `@else-if`, and `@else`:

```rtml
@if:count > 0
  <p>Positive</p>
@else
  <p>Zero or negative</p>
@endif
```

Branches may contain any valid RTML, including loops and component includes.

---

## Lists

Iterate collections with `@for`:

```rtml
@for:item in items
  <li>{item.Text}</li>
@endfor
```

* **Ranges**:

  ```rtml
  @for:i in 0..n
    <span>{i}</span>
  @endfor
  ```
* **Maps**:

  ```rtml
  @for:key,val in obj
    <p>{key}: {val}</p>
  @endfor
  ```

### Keyed items

Give items a stable identity with the `[key {expr}]` constructor to enable efficient reorders:

```rtml
@for:todo in todos
  <li [key {todo.ID}]>{todo.Text}</li>
@endfor
```

Without keys, elements may be recreated when order changes.

---

## Events

Bind DOM events with `@on:event:handler`:

```rtml
<button @on:click:increment>Increment</button>
```

You can also use the shorthand `@click:increment`—it’s **equivalent**, but `@on:` is more explicit and recommended for consistency.

### Modifiers

Append modifiers after the event name:

| Modifier  | Effect                                           |
| --------- | ------------------------------------------------ |
| `stop`    | `event.stopPropagation()` prevents bubbling      |
| `prevent` | `event.preventDefault()` blocks default behavior |
| `once`    | Removes the listener after the first call        |

Examples:

```rtml
<form @on:submit.prevent.stop:onSubmit>
<button @on:click.once:launch>Launch once</button>
```

Register handlers in Go. With Composition API, use `cmp.On("increment", fn)`; with plain `*core.HTMLComponent` use the `dom` / `events` helpers.

---

## Constructors & Template Refs

Constructors annotate elements with extra semantics. Two common patterns:

* **Template refs** — `[name]` marks the element for lookup via `GetRef("name")`.
* **List keys** — `[key {expr}]` supplies a stable key inside loops.

See **[Template Refs](./template-refs)** for details on refs and when to prefer them over DOM queries.

---

## Components, Props & Slots

Include a child component and pass props:

```rtml
@include:Card:{title:"Hello"}
```

Inside the child, read props with `@prop:name` and expose slots:

```rtml
<card>
  <h2>@prop:title</h2>
  @slot:body
    <p>Injected by parent</p>
  @endslot
</card>
```

Parents can fill named slots using `@slot:Child.slotName` from their own templates.

---

## Dynamic Components

Switch components at runtime with the `rt-is` attribute:

```rtml
<div rt-is="{current}"></div>
```

The placeholder is replaced with the component whose name matches `current`.

---

## Expressions

RTML evaluates JavaScript‑like expressions in `{}` and directive values:

```rtml
<p>{user.First + " " + user.Last}</p>
<div class="btn-{variant}"></div>
```

Expressions must produce a value. Statements (e.g. `if`, `var`) are not allowed. Only a restricted set of globals (e.g. `Math`, `Date`) is available.

---

## Plugins & Extensibility

Plugins may introduce custom variables, commands, and constructors:

* `{plugin:NAME.var}`
* `@plugin:NAME.cmd`
* `[plugin:NAME.ref]`

See the **Plugin API** for details.

---

## Security

Interpolated content is escaped by default to prevent XSS. Only render trusted data, and prefer components/slots for rich markup.

---

These building blocks cover RTML’s core syntax. Combine them with Go logic to build fully reactive components. The compiler records data dependencies so the runtime can patch only what actually changed.
