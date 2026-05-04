# Architecture

rfw v2 builds reactive web UIs entirely in **Go**. The runtime compiles to WebAssembly for the browser, while templates are written in **RTML**, an HTML-like language that binds directly to Go state. Components are struct-driven, tag-wired, and convention-resolved. No manual glue code.

---

## Composition-Based Design

v2 replaces manual component wiring with `composition.New(&MyStruct{})`. You define a Go struct with `rfw:` tags, and the framework scans, resolves templates, wires signals/stores/events/injects/lifecycle, all automatically.

```go
type Counter struct {
    composition.Component
    Count *t.Int `rfw:"signal"`
}

func (c *Counter) Inc() { c.Count.Set(c.Count.Get() + 1) }

view := composition.New(&Counter{Count: t.NewInt(0)})
```

No `dom.RegisterHandlerFunc`, no `c.Props["count"]` map lookups, no manual template loading. Tags do the wiring; convention does the rest.

---

## How composition.New Works

`composition.New(&struct{})` executes these steps in order:

1. **Validate**, Must be a pointer to a struct.
2. **Scan tags**, Parses all `rfw:` struct field tags via the `scan` package.
3. **Resolve template**, Checks `rfw:"template:path"` first, then convention (struct name → `StructName.rtml`), then panics if neither found.
4. **Create component**, Calls `core.NewHTMLComponent(name, template, nil)`.
5. **Initialize default store**, Creates or reuses the `"app"/"default"` store.
6. **Wire signals**, For each `rfw:"signal"` field: if non-nil, registers as prop; if nil, creates a zero-value signal, sets the field, and registers as prop.
7. **Wire props**, Creates `Signal[any]` entries for `rfw:"prop"` fields.
8. **Register event handlers**, For `rfw:"event:domEvent:handler"` tags, looks up the named method and calls `comp.On(handler, fn)`.
9. **Auto-discover methods**, Any exported zero-arg no-return method on the struct (excluding `Component` methods) is auto-registered as a handler.
10. **Wire lifecycle**, If `OnMount()` or `OnUnmount()` exist, they're wired to mount/unmount callbacks.
11. **Create stores**, `rfw:"store:name"` calls `comp.Store(name)`.
12. **Register hosts**, `rfw:"host:Name"` calls `AddHostComponent(name)`.
13. **Wire histories**, `rfw:"history:store:undo:redo"` retrieves the store and registers undo/redo.
14. **Wire includes**, `rfw:"include:slot"` on `*View` fields calls `AddDependency(slot, view)`.
15. **Resolve DI injects**, `rfw:"inject"` resolves from `Container()` and sets the field.
16. **Validate**, Checks signal fields are pointers, event handlers exist.

Returns `*View` (alias for `*core.HTMLComponent`).

---

## Template Resolution

Templates are resolved in this order:

1. **Explicit path**, `rfw:"template:pages/templates/custom.rtml"` loads from registered `embed.FS` instances.
2. **Convention**, No explicit tag? The struct name becomes the template name. `HomePage` → searches all registered FS for `HomePage.rtml` or `HomePage.html` (root level first, then subdirectories).
3. **Panic**, Neither found? Runtime panic.

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
    return composition.New(&components.Layout{
        Content: pages.NewHomePage(),
    })
})

router.Group("/admin", func(g *router.GroupBuilder) {
    g.Page("/dashboard", func() *t.View {
        return composition.New(&admin.Dashboard{})
    })
    g.Page("/settings", func() *t.View {
        return composition.New(&admin.Settings{})
    })
})

router.InitRouter()
```

- `Page(path, component)`, Register a single route.
- `Group(prefix, fn)`, Nest routes under a prefix with a `GroupBuilder`.
- `InitRouter()`, Starts listening for `popstate` events and navigates to the current URL.
- Routes support path params (`/user/:id`), guards, and singletons via `router.Singleton(view)`.

---

## Reactivity

### Signals, Local Reactive State

Signals are fine-grained reactive values. Use `*t.Int`, `*t.String`, `*t.Bool`, `*t.Float`, or `*t.Any`:

```go
count := t.NewInt(0)
count.Set(42)
fmt.Println(count.Get()) // 42
```

Auto-wired via `rfw:"signal"`:

```go
type Counter struct {
    composition.Component
    Count *t.Int    `rfw:"signal"`
    Name  *t.String `rfw:"signal"`
}
```

If `Count` is nil at construction, `composition.New` creates a zero-value signal automatically.

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

Auto-wired via `rfw:"store:name"`:

```go
type CartPage struct {
    composition.Component
    Cart *state.Store `rfw:"store:cart"`
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

Client components opt into host communication with `rfw:"host:Name"`:

```go
type Greeting struct {
    composition.Component
    ClientMsg *t.String `rfw:"signal"`
    HostMsg   string    `rfw:"host:GreetingHost"`
}
```

In templates, host values use `h:` prefix:

```rtml
<p>Client: @signal:ClientMsg</p>
<p>Host: {h:hostMsg}</p>
<button @on:click:h:updateTime>Refresh</button>
```

The host registers a matching component:

```go
host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
    return map[string]any{"hostMsg": "hello from server"}
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
2. **Host binary**, Standard Go build for the server.
3. **Template embedding**, Use `//go:embed` directives and `composition.RegisterFS(&fs)` in `init()`.
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

---

## DI Container

Use `.rfw:"inject"` to auto-fill dependencies from the global container:

```go
composition.Container().Register("logger", myLogger)

type Page struct {
    composition.Component
    Logger *MyLogger `rfw:"inject"`
}
```

`composition.New` resolves injectable fields via `Container().Get(key)`.

---

## v2 vs v1 Summary

| v1 | v2 |
|----|----|
| `core.NewComponent(name, tpl, map[string]any{})` | `composition.New(&MyStruct{})` |
| Manual `dom.RegisterHandlerFunc` | Auto-wired via `rfw:"signal"` + methods |
| `c.Props["count"]` map lookups | Typed fields: `c.Count.Set(x)` |
| `//go:embed` per component | `composition.RegisterFS()` globally |
| Client-only rendering | SSC required, host/client split |
| Devtools built-in | Devtools removed |
| Manual prop passing | `rfw:"include:slot"` auto-wires children |

---

## Related

- [Composition](../essentials/composition), full composition API reference
- [Signals & Effects](../essentials/signals-effects-and-watchers), reactive primitives
- [SSC](../guide/ssc), server-side computed architecture
- [State Management](../guide/state-management), stores, signals, computed values
- [Router](../guide/router), routing API