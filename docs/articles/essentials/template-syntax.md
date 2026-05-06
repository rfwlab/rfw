# Template Syntax

RTML is rfw's declarative template language. It extends HTML with **variables**, **commands** (prefixed with `@`), and **constructors** (square-bracket annotations). Templates connect the DOM to Go state and events.

---

## Building Blocks

| Construct | Syntax | Purpose |
|-----------|--------|---------|
| Variable | `{{expr}}` | Insert reactive values |
| Command | `@directive:arg` | Control flow, events, bindings |
| Constructor | `[name]` | Annotate elements (refs, keys) |

```rtml
<div [userCard]>
  <h2>{{user.Name}}</h2>
  <button @on:click:save>Save</button>
</div>
```

---

## Text Interpolation

Use `{{expr}}` to insert reactive values:

```rtml
<p>Count: {{Count.Get}}</p>
<p>Hello, {{Name.Get}}</p>
```

When the bound signal or store key changes, only the affected text node updates, no virtual DOM.

### Calling Functions

Expressions can invoke component methods:

```rtml
<time>{{FormatDate createdAt}}</time>
```

Keep expression functions pure and inexpensive.

---

## Template Expressions with @expr

`@expr:` computes a value inline from an expression:

```rtml
<p>Total: @expr:Count.Get * 2</p>
<span>@expr:Price.Get * Qty.Get</span>
```

`@expr:` is evaluated reactively, when any referenced signal changes, the output updates.

---

## Signal Bindings

`@signal:name` binds a local reactive signal:

```rtml
<p>@signal:Count</p>
<input value="@signal:Name:w">
<input type="checkbox" checked="@signal:Done:w">
<textarea>@signal:Bio:w</textarea>
```

Append `:w` on form controls for two-way binding (writes back to the signal on input). Without `:w`, the binding is read-only.

> Tip: Inside `{{}}` expressions you can reference signals by prop name: `{{Count.Get}}`. For text-only positions, `@signal:Count` is equivalent.

---

## Store Bindings

`@store:Module.Store.Key` reads from a global store:

```rtml
<p>Shared: @store:app.default.count</p>
<input value="@store:app.default.count:w">
```

Append `:w` on form controls for two-way binding. Other components reading the same store key update automatically.

---

## Event Bindings

`@on:event:handler` binds a DOM event to a handler:

```rtml
<button @on:click:Increment>+</button>
<form @on:submit.prevent:Save>...</form>
```

The handler name must match a method on the struct or a name registered via `comp.On()`.

### Modifiers

| Modifier  | Effect |
|-----------|--------|
| `.stop`    | `event.stopPropagation()` |
| `.prevent` | `event.preventDefault()` |
| `.once`    | Removes listener after first call |

Chain modifiers after the event name:

```rtml
<form @on:submit.prevent.stop:Save>
<button @on:click.once:Launch>Launch once</button>
```

---

## Conditionals

`@if`, `@else-if`, `@else`, `@endif`:

```rtml
@if:Count.Get > 0
  <p>Positive</p>
@else-if:Count.Get == 0
  <p>Zero</p>
@else
  <p>Negative</p>
@endif
```

Branches may contain any RTML content, including includes, loops, and nested conditionals.

---

## List Rendering

`@for:alias in collection` ... `@endfor`:

```rtml
@for:item in items
  <li>{{item.Text}}</li>
@endfor
```

### Range

```rtml
@for:i in 0..N.Get
  <span>{{i}}</span>
@endfor
```

### Maps

```rtml
@for:key, val in obj
  <p>{{key}}: {{val}}</p>
@endfor
```

### Keyed Items

Use `[key {{expr}}]` for stable identity and efficient reorders:

```rtml
@for:todo in todos
  <li [key {{todo.ID}}]>{{todo.Text}}</li>
@endfor
```

Without keys, elements may be recreated when order changes.

---

## Includes and Slots

### Include

`@include:component` renders a child component:

```rtml
@include:content
```

With props:

```rtml
@include:Card:{title: "Hello"}
```

### Slot

`@slot:name` ... `@endslot` defines a named outlet:

```rtml
<!-- Layout.rtml -->
<root>
  <nav>My App</nav>
  <main>@slot:content
    <p>Default content</p>
  @endslot</main>
</root>
```

Parents fill slots from their own template:

```rtml
@include:Layout
  <div slot="content">Custom content here</div>
@end
```

---

## Constructors

Square-bracket annotations in start tags:

### Template Refs

`[name]` marks an element for lookup via `GetRef("name")`:

```rtml
<div [list]></div>
```

```go
func (c *MyComp) OnMount() {
    el := c.GetRef("list")
    el.SetHTML("updated")
}
```

### List Keys

`[key {{expr}}]` gives loop items stable identity (see [List Rendering](#list-rendering)).

---

## Attribute Bindings

Attributes accept `{{expr}}` values:

```rtml
<img src="{{avatar}}" alt="{{name}}">
<div class="btn-{{variant}}"></div>
```

### Boolean Attributes

When an expression evaluates to a boolean, the attribute is present only when truthy:

```rtml
<button disabled="{{isDisabled}}">Save</button>
```

---

## @expr Directive

For computed values inline:

```rtml
<p>Doubled: @expr:Count.Get * 2</p>
<span>@expr:Price.Get * Qty.Get</span>
```

Any referenced signals trigger re-evaluation. Equivalent to a computed property but defined directly in the template.

### Ternary Expressions (Natural Syntax)

Use `if ... then ... else ...` inside `@expr:` for inline conditionals:

```rtml
<p>@expr:Count.Get > 0 then "Positive" else "Zero or negative"</p>
<span>@expr:Active.Get then "Yes" else "No"</span>
```

The condition before `then` follows the same expression rules as `@if:` — comparisons, logical operators, and signal field access all work:

```rtml
<p>@expr:Count.Get > 10 then "high" else Count.Get > 5 then "medium" else "low"</p>
```

Legacy `? :` syntax is also supported:

```rtml
<p>@expr:Count.Get > 0 ? "Positive" : "Non-positive"</p>
```

Prefer the `if ... then ... else` form — it reads naturally in RTML templates.

---

## Dynamic Components

Switch components at runtime with `rt-is`:

```rtml
<div rt-is="{{current}}"></div>
```

The placeholder is replaced with the component whose name matches `current`.

---

## Security

Interpolated content is escaped by default to prevent XSS. Only render trusted data; use components and slots for rich markup.

---

## Quick Reference

| Syntax | Purpose |
|--------|---------|
| `{{expr}}` | Text interpolation |
| `@signal:Name` | Read a signal |
| `@signal:Name:w` | Two-way signal binding |
| `@store:M.S.K` | Read a store key |
| `@store:M.S.K:w` | Two-way store binding |
| `@on:event:Handler` | Bind event |
| `@on:event.prevent.stop:Handler` | Bind event with modifiers |
| `@if:cond` ... `@endif` | Conditional |
| `@else-if:cond` |_else-if branch |
| `@else` | else branch |
| `@for:x in items` ... `@endfor` | List iteration |
| `@include:Name` | Include component |
| `@include:Name:{key: val}` | Include with props |
| `@slot:name` ... `@endslot` | Define slot |
| `@expr:expression` | Template expression |
| `@expr:cond then X else Y` | Ternary in expressions |
| `@expr:cond ? X : Y` | Legacy ternary syntax |
| `[refName]` | Template ref |
| `[key {{expr}}]` | List key |
| `rt-is="{{name}"}` | Dynamic component |