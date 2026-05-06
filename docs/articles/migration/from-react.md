# Migrating from React to rfw

You know React, JSX, hooks, context, effects. rfw replaces JavaScript with Go, the virtual DOM with fine-grained reactive updates, and component state with type-detected signal fields. This guide maps React concepts to rfw.

---

## Why Go + WASM?

React runs in JavaScript. rfw runs in Go, compiles to WASM, and requires no JS runtime on the page. Benefits:

- **Single language**, write server and client logic in Go. Share types, validators, and domain logic directly.
- **Compile-time safety**, missing signals, wrong types, and undefined handlers are caught by the Go compiler, not at runtime.
- **No virtual DOM**, rfw updates only the DOM nodes bound to changed signals. No reconciliation step.
- **SSC built-in**, Server-Side Computed rendering with WebSocket hydration is the default, not an addon.

Trade-offs: rfw's ecosystem is newer and smaller than React's. WASM initial load time is larger than a minimal React bundle. You cannot use npm packages on the client side, though you can call JavaScript via `js` interop.

---

## Mindset Shift

### React: hooks and re-renders

```jsx
function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>
}
```

Every state change triggers a re-render of the component. React reconciles the virtual DOM tree.

### rfw: signals and fine-grained updates

```go
type Counter struct {
    composition.Component
    Count *t.Int
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}
```

Signals update **only** the DOM nodes that read them. No re-render, no reconciliation. The component struct is created once and mutated in-place.

---

## Hooks Mapping

### `useState` → signal type fields

**React:**

```jsx
const [count, setCount] = useState(0)
const [name, setName] = useState('')
```

**rfw:**

```go
type MyComp struct {
    composition.Component
    Count *t.Int
    Name  *t.String
}
```

`composition.New` auto-initializes nil signal fields. Access with `.Get()` and `.Set()`.

| React | rfw type | Zero value |
|-------|----------|------------|
| `useState(0)` | `*t.Int` | `0` |
| `useState('')` | `*t.String` | `""` |
| `useState(false)` | `*t.Bool` | `false` |
| `useState(0.0)` | `*t.Float` | `0.0` |
| `useState(null)` | `*t.Any` | `nil` |

### `useEffect` → `OnMount()` / `OnUnmount()`

**React:**

```jsx
useEffect(() => {
  const timer = setInterval(() => setCount(c => c + 1), 1000)
  return () => clearInterval(timer)
}, [])
```

**rfw:**

```go
type Timer struct {
    composition.Component
    Count *t.Int
    done  chan struct{}
}

func (t *Timer) OnMount() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for {
            select {
            case <-ticker.C:
                t.Count.Set(t.Count.Get() + 1)
            case <-t.done:
                ticker.Stop()
                return
            }
        }
    }()
}

func (t *Timer) OnUnmount() {
    close(t.done)
}
```

| React | rfw | Notes |
|-------|-----|-------|
| `useEffect(fn, [])` | `OnMount()` | Auto-discovered on struct |
| `useEffect(fn, [dep])` | `state.Effect()` | Re-runs when signals change |
| Cleanup return from `useEffect` | `OnUnmount()` | Auto-discovered |
| `useEffect(fn, [val])` | `store.OnChange(key, fn)` | Store watchers |

### `useRef` → template refs

**React:**

```jsx
const inputRef = useRef(null)
useEffect(() => { inputRef.current?.focus() }, [])
// ...
<input ref={inputRef} />
```

**rfw:**

```go
type MyComp struct {
    composition.Component
    MyInput *t.Ref
}

func (c *MyComp) OnMount() {
    el := c.MyInput.Get()
    el.Call("focus")
}
```

```rtml
<input [myInput] type="text" />
```

The `[myInput]` constructor marks the element for lookup via the `*t.Ref` field.

### `useContext` → `*t.Inject[T]`

**React:**

```jsx
const ThemeContext = createContext('light')
function Button() {
  const theme = useContext(ThemeContext)
  return <button className={theme}>Click</button>
}
```

**rfw, dependency injection:**

```go
// Register globally
composition.Container().Register("theme", &ThemeService{Mode: "dark"})

// Inject into component
type Button struct {
    composition.Component
    Theme *t.Inject[*ThemeService]
}
```

`composition.New` resolves the field type from the container automatically.

For component-tree-scoped injection, use `Provide` / `Inject`:

```go
func (p *Parent) OnMount() {
    p.Provide("theme", "dark")
}

// In a child:
theme, ok := core.Inject[string](child, "theme")
```

---

## Props Flow

### React props → signal fields / composition.New

**React:**

```jsx
function Card({ title, count }) {
  return <div>{title} ({count})</div>
}
// Usage: <Card title="Hello" count={42} />
```

**rfw:**

```go
type Card struct {
    composition.Component
    Title *t.String
    Count *t.Int
}
```

Parent sets props at construction:

```go
card, err := composition.New(&Card{
    Title: t.NewString("Hello"),
    Count: t.NewInt(42),
})
if err != nil {
    log.Fatal(err)
}
```

Since signals are reactive, updating a signal from the parent propagates to the child automatically.

---

## Component Composition

### Children → `*t.View` field

**React:**

```jsx
function Layout({ children }) {
  return <div><nav>Nav</nav><main>{children}</main></div>
}
```

**rfw:**

```go
type Layout struct {
    composition.Component
    Content *t.View
}
```

Layout.rtml:

```rtml
<root>
  <nav>Nav</nav>
  <main>@include:content</main>
</root>
```

The slot name is derived from the lowercase field name (`Content` → `content`).

Parent wires the child:

```go
layout, err := composition.New(&Layout{})
if err != nil {
    log.Fatal(err)
}
layout.AddDependency("content", pageView)
```

---

## JSX → RTML Template Syntax

React uses JSX, JavaScript expressions embedded in markup. rfw uses RTML, an HTML-like template language with `@` directives.

### Conditional rendering

**React:**

```jsx
{count > 0 && <p>Positive</p>}
{count === 0 ? <p>Zero</p> : <p>Non-zero</p>}
```

**rfw:**

```rtml
@if:Count.Get > 0
  <p>Positive</p>
@endif

@if:Count.Get == 0
  <p>Zero</p>
@else
  <p>Non-zero</p>
@endif
```

### List rendering

**React:**

```jsx
{items.map(item => <li key={item.id}>{item.text}</li>)}
```

**rfw:**

```rtml
@for:item in Items
  <li [key {{item.ID}}]>{{item.Text}}</li>
@endfor
```

### Event handling

**React:**

```jsx
<button onClick={() => setCount(c => c + 1)}>+1</button>
<form onSubmit={handleSubmit}>
```

**rfw:**

```rtml
<button @on:click:Increment>+1</button>
<form @on:submit.prevent:Save>...</form>
```

Handler names reference methods on the struct:

```go
func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}
```

### Class and style bindings

**React:**

```jsx
<div className={`btn ${isActive ? 'active' : ''}`}>
</div>
```

**rfw:**

```rtml
<div class="btn @expr:IsActive.Get && 'active'"></div>
```

Or compute in Go:

```go
func (c *MyComp) ButtonClass() string {
    if c.IsActive.Get() {
        return "btn active"
    }
    return "btn"
}
```

```rtml
<div class="{{ButtonClass}}"></div>
```

---

## Routing: React Router → router.Page()

**React Router:**

```jsx
<BrowserRouter>
  <Routes>
    <Route path="/" element={<Home />} />
    <Route path="/users/:id" element={<UserProfile />} />
    <Route path="/admin" element={<AdminLayout />}>
      <Route path="dashboard" element={<Dashboard />} />
    </Route>
  </Routes>
</BrowserRouter>
```

**rfw:**

```go
func main() {
    router.Page("/", func() *composition.View {
        v, _ := composition.New(&Home{})
        return v
    })
    router.Page("/users/:id", func() *composition.View {
        v, _ := composition.New(&UserProfile{})
        return v
    })
    router.Group("/admin", func(r *router.GroupBuilder) {
        r.Page("/dashboard", func() *composition.View {
            v, _ := composition.New(&Dashboard{})
            return v
        })
    })
    router.InitRouter()
    select {}
}
```

**Route params:**

```go
func (u *UserProfile) OnMount() {
    params := u.HTMLComponent.RouteParams()
    id := params["id"] // from /users/:id
}
```

**Navigation:**

```go
router.Navigate("/users/42")
```

**Programmatic navigation in templates:**

```rtml
<a href="/users/42">View Profile</a>
```

With `router.ExposeNavigate()`, internal `<a>` clicks are intercepted for client-side navigation.

---

## SSR: Next.js → rfw SSC

Next.js offers SSR, SSG, ISR as options. rfw has SSC (Server-Side Computed) as the **required** model.

| Next.js | rfw SSC | Notes |
|---------|---------|-------|
| `getServerSideProps` | Host component handler | Server-side data provided via `h:` bindings |
| `useEffect` for hydration | Automatic WASM hydration | Browser loads WASM, attaches handlers |
| Client-side navigation | `router.Navigate()` / `<a>` interception | |
| API routes | Host component commands (`h:`) | Server-side Go functions |
| No WebSocket sync | Persistent WebSocket for `h:` updates | Live server↔client sync |

**Next.js SSR flow:**

1. Server renders React to HTML
2. Client loads React bundle
3. React hydrates the DOM
4. Client-side navigation takes over

**rfw SSC flow:**

1. Host (Go server) renders HTML with `h:` values
2. Browser receives fully rendered page
3. Browser downloads WASM bundle, hydrates DOM
4. WebSocket connects for live `h:` data sync

```go
// Host component provides server data
host.Register(host.NewHostComponent("UserHost", func(payload map[string]any) any {
    return map[string]any{"userName": "Ada"}
}))
```

```rtml
<root>
  <p>Welcome, {h:userName}</p>
  <button @on:click:h:refresh>Refresh</button>
</root>
```

---

## Context → DI Container

React context requires a Provider component that wraps consumers. rfw uses a global DI container with type-based injection.

**React:**

```jsx
const UserContext = createContext(null)

function App() {
  return (
    <UserContext.Provider value={{ name: 'Ada' }}>
      <Dashboard />
    </UserContext.Provider>
  )
}

function Dashboard() {
  const user = useContext(UserContext)
  return <p>{user.name}</p>
}
```

**rfw:**

```go
// Register
composition.Container().Register("userService", &UserService{})

// Inject
type Dashboard struct {
    composition.Component
    UserSvc *t.Inject[*UserService]
}

func (d *Dashboard) OnMount() {
    name := d.UserSvc.Get().CurrentUser()
}
```

No provider wrapping, no tree-depth propagation. The container resolves dependencies globally.

For tree-scoped values, use `Provide` / `Inject`:

```go
func (p *App) OnMount() {
    p.Provide("theme", "dark")
}

// In any descendant:
theme, _ := core.Inject[string](child, "theme")
```

---

## Common Patterns Side-by-Side

### Counter Component

**React:**

```jsx
function Counter() {
  const [count, setCount] = useState(0)
  return (
    <div>
      <p>{count}</p>
      <button onClick={() => setCount(c => c + 1)}>+1</button>
    </div>
  )
}
```

**rfw:**

```go
type Counter struct {
    composition.Component
    Count *t.Int
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}
```

```rtml
<root>
  <p>@signal:Count</p>
  <button @on:click:Increment>+1</button>
</root>
```

### Data Fetching

**React:**

```jsx
function Users() {
  const [users, setUsers] = useState([])
  useEffect(() => {
    fetch('/api/users').then(r => r.json()).then(setUsers)
  }, [])
  return (
    <ul>
      {users.map(u => <li key={u.id}>{u.name}</li>)}
    </ul>
  )
}
```

**rfw:**

```go
type Users struct {
    composition.Component
    Users *t.Any
}

func (u *Users) OnMount() {
    go func() {
        resp, _ := http.Get("/api/users")
        var users []User
        json.NewDecoder(resp.Body).Decode(&users)
        u.Users.Set(users)
    }()
}
```

```rtml
<root>
  <ul>
    @for:user in Users
      <li [key {{user.ID}}]>{{user.Name}}</li>
    @endfor
  </ul>
</root>
```

### Shared State (Redux/Context → Store)

**React:**

```jsx
const useStore = create((set) => ({
  count: 0,
  increment: () => set(s => ({ count: s.count + 1 })),
}))
```

**rfw:**

```go
s := state.NewStore("counter", state.WithModule("app"))
s.Set("count", 0)

// In a component:
type MyComp struct {
    composition.Component
    CounterStore *t.Store
}

func (m *MyComp) Increment() {
    c := m.CounterStore.Get("count").(int)
    m.CounterStore.Set("count", c+1)
}
```

```rtml
<p>@store:app.counter.count</p>
<button @on:click:Increment>+1</button>
```

---

## Quick Reference: React → rfw

| React | rfw | Notes |
|-------|-----|-------|
| `useState(val)` | `*t.Int` / `*t.String` etc (type-detected) | Typed signals |
| `useEffect(fn, [])` | `func (c *T) OnMount()` | Auto-discovered |
| `useEffect(fn, [dep])` | `state.Effect()` | Re-runs on dep change |
| `useRef()` | `*t.Ref` field + `Get()` in template | Type-based refs |
| `useContext()` | `*t.Inject[T]` field or `Provide`/`Inject` | DI container |
| JSX | RTML templates | `@` directives |
| `{condition && <El/>}` | `@if:condition` / `@endif` | Block directives |
| `{items.map(...)}` | `@for:item in items` / `@endfor` | Block directives |
| `onClick={fn}` | `@on:click:Handler` | Struct method name |
| `className={cls}` | `class="{{Method}}"` or `@expr:` | |
| `key={item.id}` | `[key {{item.ID}}]` | Constructor syntax |
| `<Child prop={val}>` | Signal fields + `composition.New` | Typed props |
| `{children}` | `@include:content` / `*t.View` field | Slots |
| `createContext` | `composition.Container()` | Global DI |
| React Router | `router.Page()` / `router.Group()` | |
| Next.js SSR | rfw SSC (required) | Host renders, WASM hydrates |
| `useReducer` | `*state.Store` | Centralized state |
| `useMemo` | Go method or `@expr:` | Computed values |
| `useCallback` | Not needed | Go methods are stable references |