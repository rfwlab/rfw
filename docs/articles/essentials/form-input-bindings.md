# Form Input Bindings

Forms capture user input. RFW simplifies handling by binding form elements directly to component fields or stores. The UI stays in sync without writing manual event listeners.

## Two-Way Store Binding

Use the `@store` directive with the `:w` suffix to bind a form control to a store key:

```rtml
<input value="@store:default.form.name:w">
```

Typing in the input updates `form.name` in the `default` module, and changes to that store reflect back into the input.

## Component Field Binding

Bind directly to a component field using `{}` with an event handler:

```rtml
<input value="{Name}" @on:input:setName>
```

```go
func (f *Form) SetName(e events.Event) {
  f.Name = e.Value.(string)
}
```

The `events.Event` passed to the handler exposes the latest value through `e.Value`.

## Checkbox and Radio

For boolean inputs, bind the `checked` attribute:

```rtml
<input type="checkbox" checked="{Done}" @on:change:toggle>
```

Handlers flip the boolean field to match the UI.

Form bindings minimize glue code so components focus on business logic instead of DOM plumbing.
