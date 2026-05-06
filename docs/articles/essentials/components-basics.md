# Components Basics

Components in rfw v2 are Go structs that embed `composition.Component` and declare reactive state via **field types** — no tags required. Templates are discovered by convention, and wiring is automatic based on type detection.

---

## Defining a Component

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

type Counter struct {
    composition.Component
    Count t.Int
}
```

Key points:

- Embed `composition.Component` (not `*core.HTMLComponent`).
- Use value types (`t.Int`, `t.String`, etc.) or pointer types (`*t.Int`, `*t.Store`) as fields.
- The struct name `Counter` resolves to `Counter.rtml` in any registered `embed.FS`.

### Creating an Instance

```go
view, err := composition.New(&Counter{})
if err != nil {
    log.Fatal(err)
}
```

`composition.New` scans field types, initializes nil pointers, discovers methods, locates the template, and returns a `*View`.

For zero-value structs without custom initialization:

```go
view, err := composition.NewFrom[Counter]()
```

---

## Type-Based Auto-Wiring

Instead of struct tags, `composition.New` detects field types and wires automatically:

| Field Type | Detection | Auto-wiring |
|---|---|---|
| `t.Int`, `t.String`, `t.Bool`, `t.Float` | Signal value types | Register as reactive prop |
| `*t.Int`, `*t.String`, etc. | Signal pointer types | Auto-init if nil, register as prop |
| `t.HInt`, `t.HString`, `t.HBool`, `t.HFloat` | Host signal types | Register as prop + host component |
| `*t.Store` | Store type | Retrieve from global manager, register on component |
| `*t.Ref` | Ref type | Allocate ref, resolve DOM node on mount |
| `*t.Inject[T]` | DI inject type | Resolve T from DI container by lowercase field name |
| `*t.History` | History type | Bind to component's first store for undo/redo |
| `*t.View` | Include type | AddDependency(lowercase field name, view) |
| `*t.Slice[T]`, `*t.Map[K,V]` | Collection signal types | Register as reactive prop |
| `t.Prop[T]` | Prop type | Create reactive prop |

### Signals

Signal fields are registered as reactive props. If the field is a nil pointer, it's auto-initialized with a zero value:

```go
type TodoApp struct {
    composition.Component
    Count t.Int      // value type — works directly
    Label *t.String  // pointer type — auto-initialized if nil
    Done  t.Bool
}
```

Available signal types:

| Type | Zero value |
|------|-----------|
| `t.Int` | `0` |
| `t.String` | `""` |
| `t.Bool` | `false` |
| `t.Float` | `0.0` |

### Stores

```go
type Settings struct {
    composition.Component
    LocalStore *t.Store
}
```

`composition.New` retrieves the store from the global manager and registers it on the component. Access in templates with `@store:app.Settings.LocalStore.key`.

### Injects

```go
type Dashboard struct {
    composition.Component
    Logger *t.Inject[Logger]
}
```

Register providers before calling `composition.New`:

```go
composition.Container().Provide("logger", &MyLogger{})
```

`composition.New` resolves `*t.Inject[T]` fields from the DI container using the lowercase field name as the key.

### Refs

```go
type Form struct {
    composition.Component
    Input *t.Ref
}

func (f *Form) OnMount() {
    // Input is automatically resolved from the DOM
    val := f.Input.Get()
    val.Call("focus")
}
```

Refs are allocated during `composition.New` and resolved from the DOM on mount. Use `[name]` in RTML to mark elements:

```rtml
<input [nameInput]>
```

### History (Undo/Redo)

```go
type Editor struct {
    composition.Component
    Doc *t.Store
    Hist *t.History
}
```

`*t.History` fields are automatically bound to the component's first store. Call `Snapshot()`, `Undo()`, `Redo()`:

```go
func (e *Editor) OnMount() {
    e.Hist.Snapshot() // save current state
}

func (e *Editor) Undo() { e.Hist.Undo() }
func (e *Editor) Redo() { e.Hist.Redo() }
```

### Includes

```go
type Layout struct {
    composition.Component
    Content *t.View
}
```

`*t.View` fields are auto-wired as includes using the lowercase field name as the slot:

```rtml
<root>
  <nav>My App</nav>
  <main>@include:content</main>
</root>
```

---

## Template Convention

By default, `composition.New` finds a template matching the struct name:

| Struct | Template searched |
|--------|-------------------|
| `HomePage` | `HomePage.rtml` or `HomePage.html` |
| `Counter` | `Counter.rtml` or `Counter.html` |

It searches root-level first, then recursively through subdirectories, across all registered `embed.FS` instances.

### Overriding the Template

Define a `Template() string` method on your struct:

```go
type HomePage struct {
    composition.Component
    Count t.Int
}

func (h *HomePage) Template() string {
    return "<root><h1>@signal:Count</h1></root>"
}
```

---

## Lifecycle Methods

`composition.New` auto-discovers no-argument exported methods named `OnMount` and `OnUnmount`:

```go
type Tracker struct {
    composition.Component
    Count t.Int
}

func (t *Tracker) OnMount() {
    t.Count.Set(0)
}

func (t *Tracker) OnUnmount() {
    // cleanup subscriptions, timers, etc.
}
```

No registration needed, just define the methods. On mount, refs are also resolved from the DOM.

---

## Event Handlers

Any exported no-argument, no-return method (excluding `OnMount`, `OnUnmount`, and Component methods like `On`, `Prop`, `Store`, `History`, `Unwrap`) is auto-registered as an event handler:

```go
type App struct {
    composition.Component
    Count t.Int
}

func (a *App) Increment() { a.Count.Set(a.Count.Get() + 1) }
func (a *App) Decrement() { a.Count.Set(a.Count.Get() - 1) }
```

In RTML:

```rtml
<button @on:click:Increment>+1</button>
<button @on:click:Decrement>-1</button>
```

---

## Full Example

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
    "github.com/rfwlab/rfw/v2/types"
)

type App struct {
    composition.Component
    Count   t.Int
    Sidebar *types.View
}

func (a *App) OnMount() {
    a.Count.Set(0)
}

func (a *App) Increment() {
    a.Count.Set(a.Count.Get() + 1)
}
```

`App.rtml`:

```rtml
<root>
  <h1>Counter: @signal:Count</h1>
  <button @on:click:Increment>+1</button>
  @include:sidebar
</root>
```

Wiring it up:

```go
// main.go
func main() {
    router.Page("/", func() *types.View {
        view, err := composition.NewFrom[App]()
        if err != nil {
            log.Fatal(err)
        }
        return view
    })
    router.InitRouter()
    select {}
}
```

---

## Host Component Types (SSC)

Use `t.HInt`, `t.HString`, `t.HBool`, `t.HFloat` for server-side computed values:

```go
type VisitPage struct {
    composition.Component
    Visit t.HInt
}
```

These host signal types are both reactive (like `t.Int`) and automatically register as host component bindings. In the template:

```rtml
<root>
  <p>Visits: @signal:Visit</p>
  <button @on:click:h:UpdateVisits>refresh</button>
</root>
```

---

## See Also

- [Composition](./composition), how `composition.New` works internally
- [Template Syntax](./template-syntax), RTML directives reference
- [Signals & Effects](./signals-effects-and-watchers), reactive primitives