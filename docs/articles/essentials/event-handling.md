# Event Handling

Interactivity in **rfw** is driven by events. Templates register listeners with directives, and Go code defines the handlers. rfw wires them together and cleans up automatically.

---

## DOM Events

Attach a handler to a browser event with the `@on:` directive:

```rtml
<button @on:click:save>Save</button>
```

The matching Go method receives an `events.Event` value:

```go
func (c *Form) Save(e events.Event) {
  // handle submission
}
```

### `@click` vs `@on:click`

You may also see the shorthand `@click:save`:

```rtml
<button @click:save>Save</button>
```

This is equivalent to `@on:click:save`. Both work, but `@on:click` is more explicit and recommended for readability and consistency.

---

## Event Modifiers

Modifiers adjust how listeners behave. Append them after the event name:

```rtml
<form @on:submit.prevent.once:onSubmit>
```

Supported modifiers:

| Modifier  | Description                                                          |
| --------- | -------------------------------------------------------------------- |
| `stop`    | Calls `event.stopPropagation()` to prevent bubbling.                 |
| `prevent` | Calls `event.preventDefault()` to stop the browser’s default action. |
| `once`    | Removes the listener after the first invocation.                     |

Examples:

```rtml
<button @on:click.stop.prevent:save>Save</button>
<button @on:click.once:load>Load once</button>
```

---

## Registering from Go

Handlers can also be attached programmatically using the `events` package:

```go
doc := dom.Doc()
stop := events.OnClick(doc.ByID("save").Value, func(evt js.Value) {
    // handle click
})
defer stop()
```

This is useful for dynamically created elements or low-level control.

---

## Store Events

Reactive stores emit events whenever their values update. Components can subscribe with **watchers** or **computed properties**. See the *Watchers* section for details. This lets components react to global state changes without manual DOM listeners.

---

## Custom Events

Components may define their own events by exposing channels or calling callbacks passed via props. Event emitters are plain Go functions, keeping the API lightweight.

---

## Summary

* Use `@on:event:handler` in templates to bind DOM events.
* `@click:handler` is shorthand for `@on:click:handler`, but `@on:` is preferred for clarity.
* Event modifiers like `.stop`, `.prevent`, and `.once` give finer control.
* You can also attach events programmatically or react to store changes.

Event handling in rfw unifies user input, state changes, and custom logic under a clear, declarative syntax.
