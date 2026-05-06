# composition

```go
import "github.com/rfwlab/rfw/v2/composition"
```

Package for struct-based component creation, DI, template resolution, and DOM composition helpers. All wiring is type-based — no struct tags required.

---

## Core Functions

| Function | Description |
| --- | --- |
| `New(v any) (*View, error)` | Creates a View from a struct pointer. Scans field types to auto-wire signals, stores, refs, injects, histories, host bindings, includes, and template resolution. Auto-discovers `OnMount`/`OnUnmount` methods. Template found by `Template()` method or struct name convention (`StructName.rtml`). Returns error instead of panicking. |
| `NewFrom[T any]() (*View, error)` | Generic factory. Creates a zero-value `T`, then calls `New` on it. |
| `NewRaw(name string, tpl []byte, props map[string]any) *View` | Creates a View from a raw template. No type scanning. Use for layout/wrapper components. |
| `RegisterFS(fs *embed.FS)` | Registers an `embed.FS` for template resolution. Call in `init()` or at package level. Multiple FS instances are searched in registration order. |
| `Container() *fndi.Container` | Returns the default DI container used by `New` for `*t.Inject[T]` resolution. |
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

## Type-Based Auto-Wiring

`composition.New` detects field types and wires automatically. No `rfw:` tags are used.

| Field Type | Auto-wiring |
| --- | --- |
| `t.Int`, `t.String`, `t.Bool`, `t.Float` (value types) | Register as reactive prop via `field.Addr()` |
| `*t.Int`, `*t.String`, etc. (pointer types) | Auto-init if nil, register as prop |
| `t.HInt`, `t.HString`, `t.HBool`, `t.HFloat` | Register as prop + host component binding |
| `*t.Slice[T]`, `*t.Map[K,V]` | Register as reactive prop |
| `*t.Store` | Retrieve from global manager, register on component |
| `*t.Ref` | Allocate ref, resolve DOM node on mount |
| `*t.Inject[T]` | Resolve T from DI container by lowercase field name |
| `*t.History` | Bind to component's first store for undo/redo |
| `*t.View` | `AddDependency(lowercase field name, view)` |
| `t.Prop[T]` | Create reactive prop |
| Exported `func()` methods | Register as event handlers (excluding `OnMount`, `OnUnmount`, Component methods) |
| `Template() string` method | Override template resolution |

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