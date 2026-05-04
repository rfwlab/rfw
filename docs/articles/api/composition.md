# composition

```go
import "github.com/rfwlab/rfw/v2/composition"
```

Package for struct-based component creation, DI, template resolution, and DOM composition helpers.

---

## Core Functions

| Function | Description |
| --- | --- |
| `New(v any) *View` | Creates a View from a struct pointer. Scans `rfw` tags to auto-wire signals, stores, events, includes, injects, hosts, refs, FSMs, histories, and template resolution. Auto-discovers `OnMount`/`OnUnmount` methods. Template found by `rfw:"template:path"` tag or struct name convention (`StructName.rtml`). |
| `NewFrom[T any]() *View` | Generic factory. Creates a zero-value `T`, then calls `New` on it. |
| `NewRaw(name string, tpl []byte, props map[string]any) *View` | Creates a View from a raw template. No tag scanning. Use for layout/wrapper components. |
| `RegisterFS(fs *embed.FS)` | Registers an `embed.FS` for template resolution. Call in `init()` or at package level. Multiple FS instances are searched in registration order. |
| `Container() *fndi.Container` | Returns the default DI container used by `New` for `rfw:"inject"` resolution. |
| `SetDevMode(v bool)` | Enables or disables development mode (verbose logging, validation). |

---

## Component Type

`Component` wraps `*core.HTMLComponent` with composition helpers. Created via `Wrap` or internally by `New`.

| Method | Description |
| --- | --- |
| `On(name string, fn func())` | Registers a DOM event handler by name. |
| `Prop(key string, sig signalAny)` | Associates a reactive signal with the component under `key`. |
| `Store(name string, opts ...state.StoreOption) *state.Store` | Creates or retrieves a namespaced store for this component. |
| `History(s *state.Store, undo, redo string)` | Registers undo/redo handlers for a store. |
| `Unwrap() *core.HTMLComponent` | Returns the underlying `*core.HTMLComponent`. |

---

## FromProp

```go
func FromProp[T any](c *Component, key string, def T) *Signal[T]
```

Creates a signal from a component prop. If the prop exists and is a `*Signal[T]`, returns it. Otherwise returns a new signal with `def` as initial value.

---

## Type Aliases

Re-exports from the `types` package for convenience:

| Alias | Underlying Type |
| --- | --- |
| `Int` | `state.Signal[int]` |
| `String` | `state.Signal[string]` |
| `Bool` | `state.Signal[bool]` |
| `Float` | `state.Signal[float64]` |
| `Any` | `state.Signal[any]` |
| `Store` | `state.Store` |
| `View` | `core.HTMLComponent` |
| `Comp` | `core.Component` |
| `Viewer` | `types.Viewer` (interface) |

---

## Signal Constructors

| Function | Signature | Description |
| --- | --- | --- |
| `NewInt` | `func(v int) *Int` | Creates a signal initialized to `v`. |
| `NewString` | `func(v string) *String` | Creates a signal initialized to `v`. |
| `NewBool` | `func(v bool) *Bool` | Creates a signal initialized to `v`. |
| `NewFloat` | `func(v float64) *Float` | Creates a signal initialized to `v`. |
| `NewAny` | `func(v any) *Any` | Creates a signal initialized to `v`. |

---

## Tag Reference

`composition.New` scans struct fields for `rfw` tags:

| Tag | Format | Description |
| --- | --- | --- |
| `rfw:"signal"` | `rfw:"signal"` or `rfw:"signal:name"` | Reactive signal field. Auto-creates zero-value signal if nil. |
| `rfw:"store:name"` | `rfw:"store:counter"` | Creates a component-scoped store. |
| `rfw:"inject"` | `rfw:"inject"` or `rfw:"inject:key"` | DI injection from the container. |
| `rfw:"include:slot"` | `rfw:"include:sidebar"` | Includes a child View in the named slot. Field must be `*View`. |
| `rfw:"template:path"` | `rfw:"template:templates/foo.rtml"` | Explicit template path in registered FS. |
| `rfw:"host:name"` | `rfw:"host:modal"` | Registers a host component. |
| `rfw:"event:click:Handler"` | `rfw:"event:click:HandleSave"` | Binds a DOM event to a method on the struct. |
| `rfw:"prop:name"` | `rfw:"prop:title"` | Declares a prop signal. |
| `rfw:"ref:name"` | `rfw:"ref:inputEl"` | Template ref for DOM lookup. |
| `rfw:"fsm:definition"` | `rfw:"fsm:toggle"` | Finite state machine definition. |
| `rfw:"history:store:undo:redo"` | `rfw:"history:counter:Undo:Redo"` | Undo/redo history on a store. |

---

## Element Builders

Programmatic DOM construction:

| Function | Returns | Description |
| --- | --- | --- |
| `Div()` | `*divNode` | Creates a `<div>` builder. |
| `A()` | `*anchorNode` | Creates an `<a>` builder. |
| `Span()` | `*spanNode` | Creates a `<span>` builder. |
| `Button()` | `*buttonNode` | Creates a `<button>` builder. |
| `H(level int)` | `*headingNode` | Creates `<h1>`..`<h6>` (clamped 1-6). |

All builders share chainable methods: `Class(name)`, `Classes(names...)`, `Style(prop, val)`, `Styles(props...)`, `Text(t)`, `Group(g *Elements)`. Each implements the `Node` interface via `Element() dom.Element`.

`anchorNode` adds `Attr(name, val)` and `Href(h)`.

---

## Composition Helpers

| Function | Description |
| --- | --- |
| `Wrap(c *core.HTMLComponent) *Component` | Wraps an HTMLComponent into a composition.Component. |
| `Group(nodes ...Node) *Elements` | Collects nodes into a group for bulk operations. |
| `NewGroup() *Elements` | Creates an empty group. |
| `Bind(selector string, fn func(El))` | Selects first matching element and applies `fn`. |
| `BindEl(el dom.Element, fn func(El))` | Like `Bind` but with a direct element reference. |
| `For(selector string, fn func() Node)` | Repeatedly calls `fn`, appending nodes to matched element until `fn` returns nil. |

---

## Elements Methods

`*Elements` supports bulk operations on grouped elements:

`ForEach`, `AddClass`, `RemoveClass`, `ToggleClass`, `SetAttr`, `SetStyle`, `SetText`, `SetHTML`, `Group`.