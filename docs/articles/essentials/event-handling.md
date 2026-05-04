# Event Handling

Interactivity in **rfw v2** is driven by events. Templates register listeners with directives, and Go code defines the handlers. `composition.New` auto-wires them together.

---

## Auto-Discovered Handlers

Exported no-argument methods on a struct are automatically registered as event handlers by `composition.New`:

```go
type Form struct {
    composition.Component
    Count *t.Int `rfw:"signal"`
}

func (f *Form) Save() {
    f.Count.Set(f.Count.Get() + 1)
}
```

The method name (`Save`) becomes the handler name. Wire it in your template:

```rtml
<button @on:click:Save>Save</button>
```

Any exported method with no arguments and no return value is auto-discovered and registered as a handler. No tags required.

---

## Tag-Based Registration

For explicit control, use the `rfw:"event"` tag on a field:

```go
type Form struct {
    composition.Component
    _ struct{} `rfw:"event:click:Save"`
}
```

This registers the `Save` method for the `click` DOM event. You can also specify modifiers:

```go
_ struct{} `rfw:"event:submit.prevent:HandleForm"`
```

---

## @on: Directive

The primary template syntax for event binding is `@on:event:Handler`:

```rtml
<button @on:click:Save>Save</button>
<form @on:submit.prevent:HandleForm>...</form>
```

`composition.New` reads the handler name, finds the corresponding method on the struct, and wires the DOM event.

---

## Event Modifiers

Modifiers adjust how listeners behave. Append them after the event name, separated by dots:

```rtml
<form @on:submit.prevent.once:HandleForm>
```

| Modifier  | Description                                                          |
| --------- | -------------------------------------------------------------------- |
| `.stop`   | Calls `event.stopPropagation()`, prevents bubbling.                |
| `.prevent` | Calls `event.preventDefault()`, stops browser default action.     |
| `.once`   | Removes the listener after the first invocation.                     |

Examples:

```rtml
<button @on:click.stop.prevent:Save>Save</button>
<button @on:click.once:Load>Load once</button>
```

---

## CamelCase Dataset

When rfw registers event handlers, it creates DOM event listeners that reference handlers by name via `data-rfw-*` attributes. Handler names use **CamelCase**, no kebab-case conversion:

```rtml
<button @on:click:handleSubmit>Submit</button>
```

The handler method must be `func (c *T) handleSubmit()`, matching the exact name used in the template.

---

## Full Example

```go
package main

import (
    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

type Counter struct {
    composition.Component
    Count *t.Int `rfw:"signal"`
}

//go:embed templates/counter.rtml
var templates embed.FS

func init() {
    composition.RegisterFS(&templates)
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}

func (c *Counter) Decrement() {
    c.Count.Set(c.Count.Get() - 1)
}

func (c *Counter) Reset() {
    c.Count.Set(0)
}
```

```rtml
<div>
  <p>Count: @signal:Count</p>
  <button @on:click:Increment>+</button>
  <button @on:click:Decrement>-</button>
  <button @on:click:Reset>Reset</button>
</div>
```

Each exported no-arg method (`Increment`, `Decrement`, `Reset`) is auto-discovered and available by name in the template.

---

## Registering from Go

For dynamically created elements or low-level control, use the `events` package directly:

```go
import "github.com/rfwlab/rfw/v2/events"

stop := events.OnClick(target, func(evt js.Value) {
    // handle click
})
defer stop()
```

This bypasses the composition system entirely, use it only when you need programmatic control outside the template.

---

## Summary

* **Auto-discovered**: exported no-arg methods on the struct are handlers.
* **Tag-based**: `rfw:"event:click:Handler"` for explicit registration.
* **Template**: `@on:event:Handler` to bind DOM events.
* **Modifiers**: `.stop`, `.prevent`, `.once` for finer control.
* **Go API**: `events.On*` for low-level programmatic listeners.