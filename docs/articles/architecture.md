# Architecture

rfw v2 builds reactive web UIs entirely in **Go**. The runtime compiles to WebAssembly for the browser, while templates are written in **RTML**, an HTML-like language that binds directly to Go state. Components are struct-driven, type-wired, and convention-resolved. No manual glue code.

---

## Composition-Based Design

v2 replaces manual component wiring with `composition.New(&MyStruct{})`. You define a Go struct with typed fields, and the framework scans, resolves templates, wires signals/stores/events/injects/lifecycle — all automatically.

```go
type Counter struct {
    composition.Component
    Count t.Int
}

func (c *Counter) Inc() { c.Count.Set(c.Count.Get() + 1) }

view, err := composition.New(&Counter{Count: *t.NewInt(0)})
```

No `dom.RegisterHandlerFunc`, no `c.Props["count"]` map lookups, no manual template loading. Field types determine wiring; convention does the rest.

---

## How composition.New Works

`composition.New(&struct{})` executes these steps in order:

1. **Validate**, must be a pointer to a struct.
2. **Scan field types**, the `scan` package inspects field types (not tags) to detect signals, stores, refs, injects, histories, host types, and includes.
3. **Resolve template**, checks for a `Template() string` method first, then convention (struct name → `StructName.rtml` or `StructName.html`). Returns error if neither found.
4. **Create component**, calls `core.NewHTMLComponent(name, template, nil)`.
5. **Initialize default store**, creates or reuses the `"app"/"default"` store.
6. **Wire signals**, for each signal-type field: if nil pointer, creates a zero-value signal and sets the field; registers as prop.
7. **Wire host signals**, `t.HInt`, `t.HString`, etc. are registered as both props and host component bindings.
8. **Wire stores**, `*t.Store` fields get the store from the global manager.
9. **Wire histories**, `*t.History` fields are bound to the component's first store.
10. **Wire includes**, `*t.View` fields call `AddDependency(lowercase field name, view)`.
11. **Wire injects**, `*t.Inject[T]` fields are resolved from the DI container.
12. **Wire refs**, `*t.Ref` fields are allocated and resolved from the DOM on mount.
13. **Auto-discover methods**, exported zero-arg no-return methods are auto-registered as event handlers (excluding `OnMount`/`OnUnmount` and Component methods).
14. **Wire lifecycle**, `OnMount` and `OnUnmount` are registered; refs are resolved before `OnMount`.

Returns `(*View, error)`. Returns a descriptive error instead of panicking on failure.

---

## Template Resolution

Templates are resolved in this order:

1. **Template() method**, if the struct implements `Template() string`, that string is used directly.
2. **Convention**, the struct name becomes the template name. `HomePage` → searches all registered FS for `HomePage.rtml` or `HomePage.html` (root level first, then subdirectories).
3. **Error**, if neither found, returns an error.

Register your template FS in `init()`:

```go
//go:embed pages/templates components/templates
var templates embed.FS

func init() {
    composition.RegisterFS(&templates)
}
```

---

## Router

The v2 router uses `Page()` and `Group()` for route definition:

```go
router.Page("/", func() *t.View {
    view, _ := composition.New(&components.Layout{
        Content: pages.NewHomePage(),
    })
    return view
})

router.Group("/admin", func(g *router.GroupBuilder) {
    g.Page("/dashboard", func() *t.View {
        v, _ := composition.New(&admin.Dashboard{})
        return v
    })
    g.Page("/settings", func() *t.View {
        v, _ := composition.New(&admin.Settings{})
        return v
    })
})

router.InitRouter()
```

- `Page(path, component)`, register a single route.
- `Group(prefix, fn)`, nest routes under a prefix with a `GroupBuilder`.
- `InitRouter()`, starts listening for `popstate` events and navigates to the current URL.
- Routes support path params (`/user/:id`), guards, and singletons via `router.Singleton(view)`.

---

## Reactivity

### Signals, Local Reactive State

Signals are fine-grained reactive values. Use `t.Int`, `t.String`, `t.Bool`, `t.Float`, or `t.Any`:

```go
count := t.NewInt(0)
count.Set(42)
fmt.Println(count.Get()) // 42
```

Auto-wired by type detection:

```go
type Counter struct {
    composition.Component
    Count t.Int      // value type — auto-wired
    Name  *t.String  // pointer type — auto-initialized if nil
}
```

If `Name` is nil at construction, `composition.New` creates a zero-value signal automatically.

In templates:

```rtml
<span>@signal:Count</span>
<input value="@signal:Name:w">
```

- `@signal:Name`, read-only binding
- `@signal:Name:w`, two-way binding (writes back on input)

### Stores, Global State

Stores are namespaced key-value maps shared across components:

```go
s := state.NewStore("cart", state.WithModule("app"), state.WithHistory(10))
s.Set("count", 0)
s.Set("name", "rfw")
```

Auto-wired by type:

```go
type CartPage struct {
    composition.Component
    Cart *t.Store
}
```

In templates:

```rtml
<p>Items: @store:app.cart.count</p>
<input value="@store:app.cart.name:w">
```

Format: `@store:module.store.key`, module defaults to `app`, so `@store:app.default.count` is the default store.

### Computed Values, @expr:

Inline computed expressions re-evaluate when dependencies change:

```rtml
<p>Double: @expr:Count.Get * 2</p>
```

Or define computed stores in Go:

```go
state.Map2(s, "fullName", "first", "last", func(first, last string) string {
    return first + " " + last
})
```

---

## RTML Template Directives

RTML connects Go state to the DOM through directives:

| Directive | Purpose | Example |
|-----------|---------|---------|
| `@signal:Name` | Read a signal | `@signal:Count` |
| `@signal:Name:w` | Two-way signal binding | `value="@signal:Name:w"` |
| `@store:m.s.k` | Read a store key | `@store:app.cart.total` |
| `@store:m.s.k:w` | Two-way store binding | `value="@store:app.cart.name:w"` |
| `@expr:` | Computed expression | `@expr:Count.Get * 2` |
| `@on:click:handler` | DOM event → Go method | `@on:click:Increment` |
| `@include:slot` | Inject child view | `@include:content` |
| `@if:cond` / `@endif` | Conditional rendering | `@if:signal:Count == "3"` |
| `@for:item in signal:Items` / `@endfor` | List rendering | Loop over signals |

---

## Host/Client SSC Split

SSC (Server-Side Computed) is **required in v2**. The architecture splits into two processes:

- **Host**, Go server that renders HTML, runs privileged logic, and serves the Wasm bundle.
- **Client**, Wasm bundle that hydrates server HTML, handles local reactivity, and syncs with the host over WebSocket.

Client components declare host bindings with host signal types (`t.HInt`, `t.HString`, etc.):

```go
type Greeting struct {
    composition.Component
    ClientMsg t.String
    Visit     t.HInt
}
```

In templates, host values use `h:` prefix:

```rtml
<p>Client: @signal:ClientMsg</p>
<p>Host: {h:visit}</p>
<button @on:click:h:updateTime>Refresh</button>
```

The host registers a matching component:

```go
host.Register(host.NewHostComponent("Visit", func(_ map[string]any) any {
    return map[string]any{"visit": 0}
}))
```

The SSC server serves static files and handles WebSocket connections:

```go
sscSrv := ssc.NewSSCServer(":8080", "client")
sscSrv.ListenAndServe()
```

---

## Build Pipeline

rfw v2 compiles your Go code to WebAssembly and embeds templates via `embed.FS`:

1. **Go → WASM**, `GOOS=js GOARCH=wasm go build -o app.wasm` for the client.
2. **Host binary**, standard Go build for the server.
3. **Template embedding**, use `//go:embed` directives and `composition.RegisterFS(&fs)` in `init()`.
4. **rfw CLI**, `rfw build` produces both `build/client/` (Wasm + assets) and `build/host/` (server binary). `rfw dev` watches and rebuilds on changes.

### Project Structure

```
myapp/
├── main.go              // client entry, router, RegisterFS
├── components/
│   ├── layout.go
│   └── templates/
│       └── Layout.rtml
├── pages/
│   ├── home.go
│   └── templates/
│       └── HomePage.rtml
├── host/
│   └── main.go          // host server
└── rfw.json              // build config
```

---

## Component Lifecycle

Components can implement `OnMount()` and `OnUnmount()`, zero-arg methods auto-wired by `composition.New`:

```go
func (c *Counter) OnMount() {
    c.Count.Set(0)
}

func (c *Counter) OnUnmount() {
    c.Count.Set(0)
}
```

Refs (`*t.Ref` fields) are resolved from the DOM before `OnMount` runs.

---

## DI Container

Use `*t.Inject[T]` fields to auto-fill dependencies from the global container:

```go
composition.Container().Provide("logger", myLogger)

type Page struct {
    composition.Component
    Logger *t.Inject[Logger]
}
```

`composition.New` resolves injectable fields via `Container().Get("logger")` using the lowercase field name as the key.

---

## v2 vs v1 Summary

| v1 | v2 |
|----|----|
| `core.NewComponent(name, tpl, map[string]any{})` | `composition.New(&MyStruct{})` |
| Manual `dom.RegisterHandlerFunc` | Auto-wired via type detection + methods |
| `c.Props["count"]` map lookups | Typed fields: `c.Count.Set(x)` |
| `//go:embed` per component | `composition.RegisterFS()` globally |
| Client-only rendering | SSC required, host/client split |
| Devtools built-in | Devtools removed |
| Manual prop passing | `*t.View` fields auto-wire children |
| `rfw:` struct tags | Type-based detection (no tags needed) |

---

## Related

- [Composition](../essentials/composition), full composition API reference
- [Signals & Effects](../essentials/signals-effects-and-watchers), reactive primitives
- [SSC](../guide/ssc), server-side computed architecture
- [State Management](../guide/state-management), stores, signals, computed values
- [Router](../guide/router), routing API