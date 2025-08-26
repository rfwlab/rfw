# Event Handling

Interactivity in RFW is driven by events. Templates register listeners with the `@on:` directive while Go code defines the handlers. RFW wires them together and cleans up automatically.

## DOM Events

Attach a handler to a browser event by specifying `@on:event:method`:

```rtml
<button @on:click:save>Save</button>
```

The corresponding method on the component receives an `events.Event` value:

```go
func (c *Form) Save(e events.Event) {
  // handle submission
}
```

Modifiers such as `.prevent`, `.stop`, or `.once` may be appended after the event name to control default behavior or limit invocation:

```rtml
<form @on:submit.prevent.once:onSubmit>
```

Handlers can also be registered from Go code using the `events` package:

```go
stop := events.OnClick(dom.ByID("save"), func(evt js.Value) {
        // handle click
})
defer stop()
```

This approach is useful when listeners need to be attached dynamically.

## Store Events

Reactive stores emit change events whenever their values update. Components can listen by registering a watcher or computed property; see the Watchers section for details. This allows components to react to global state changes without manual DOM listeners.

## Custom Events

Components may expose their own events by embedding channels or calling parent callbacks passed via props. Emitters remain plain Go, keeping the API lightweight.

Event handling unifies user input and state changes under a consistent, declarative syntax.
