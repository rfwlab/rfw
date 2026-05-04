# Components Basics

Components in rfw v2 are Go structs that embed `composition.Component` and declare reactive state via struct tags. Templates are discovered by convention, and wiring is automatic.

---

## Defining a Component

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
)

type Counter struct {
    composition.Component
    Count *composition.Int `rfw:"signal"`
}
```

Key points:

- Embed `composition.Component` (not `*core.HTMLComponent`).
- Tag pointer-to-signal fields with `rfw:"signal"`.
- The struct name `Counter` resolves to `Counter.rtml` in any registered `embed.FS`.

### Creating an Instance

```go
view := composition.New(&Counter{})
```

`composition.New` scans tags, initializes nil signal fields, discovers methods, locates the template, and returns a `*View`.

For zero-value structs without custom initialization:

```go
view := composition.NewFrom[Counter]()
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

Use `rfw:"template:path"` on a blank struct field:

```go
type HomePage struct {
    composition.Component
    _ struct{} `rfw:"template:pages/home.rtml"`
    Count *composition.Int `rfw:"signal"`
}
```

---

## Tag-Driven Wiring

### Signals

`rfw:"signal"` registers a reactive signal as a template prop. If the field is `nil`, `composition.New` auto-initializes it with a zero-value signal:

```go
type TodoApp struct {
    composition.Component
    Count *composition.Int   `rfw:"signal"`
    Label *composition.String `rfw:"signal"`
    Done  *composition.Bool   `rfw:"signal"`
}
```

Available signal types:

| Type | Constructor | Zero value |
|------|------------|-------------|
| `*composition.Int` | `composition.NewInt(0)` | `0` |
| `*composition.String` | `composition.NewString("")` | `""` |
| `*composition.Bool` | `composition.NewBool(false)` | `false` |
| `*composition.Float` | `composition.NewFloat(0.0)` | `0.0` |

These are re-exports of `*state.Signal[T]`, see [Reactivity Fundamentals](./reactivity-fundamentals).

### Stores

`rfw:"store"` creates a component-scoped store:

```go
type Settings struct {
    composition.Component
    LocalStore *composition.Store `rfw:"store"`
}
```

Access in templates with `@store:Settings.LocalStore.key`.

### Inject

`rfw:"inject"` resolves a dependency from the DI container. Without a key, it defaults to the field name:

```go
type Dashboard struct {
    composition.Component
    UserService *UserService `rfw:"inject"`         // key defaults to "UserService"
    Logger      *Logger       `rfw:"inject:appLog"`  // explicit key
}
```

Register providers before calling `composition.New`:

```go
composition.Container().Register("UserService", &myService{})
```

### Include

`rfw:"include:slotname"` wires a child `*View` into a named slot:

```go
type Layout struct {
    composition.Component
    Content *types.View `rfw:"include:content"`
}
```

In the parent:

```go
layout := composition.New(&Layout{})
```

The `AddDependency("content", childView)` call happens automatically. In the template, `@include:content` renders the child at that position.

### Host

`rfw:"host"` registers a host component by name:

```go
type Page struct {
    composition.Component
    _ string `rfw:"host:Analytics"`
}
```

---

## Lifecycle Methods

`composition.New` auto-discovers no-argument exported methods named `OnMount` and `OnUnmount`:

```go
type Tracker struct {
    composition.Component
    Count *composition.Int `rfw:"signal"`
}

func (t *Tracker) OnMount() {
    t.Count.Set(0)
}

func (t *Tracker) OnUnmount() {
    // cleanup subscriptions, timers, etc.
}
```

No registration needed, just define the methods.

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
    Count   *composition.Int  `rfw:"signal"`
    Sidebar *types.View       `rfw:"include:sidebar"`
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
    router.Page("/", composition.NewFrom[App]())
    router.InitRouter()
    select {}
}
```

---

## Layout Pattern

A common pattern is a layout component with a content slot:

```go
type Layout struct {
    composition.Component
    Content *types.View `rfw:"include:content"`
}
```

`Layout.rtml`:

```rtml
<root>
  <nav>My App</nav>
  <main>@include:content</main>
</root>
```

Parent pages compose with the layout:

```go
type Home struct {
    composition.Component
    Count *composition.Int `rfw:"signal"`
    Layout *types.View
}

func (h *Home) OnMount() {
    h.Layout = composition.New(&Layout{})
    // wire Home into Layout's content slot
    // ...
}
```

---

## See Also

- [Composition](./composition), how `composition.New` works internally
- [Template Syntax](./template-syntax), RTML directives reference
- [Signals & Effects](./signals-effects-and-watchers), reactive primitives