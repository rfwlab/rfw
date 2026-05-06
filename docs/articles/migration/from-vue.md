# Migrating from Vue.js to rfw

You know Vue, reactive data, SFCs, computed properties, v-model, slots, Vuex/Pinia. rfw borrows many of the same ideas but implements them in Go, compiled to WebAssembly, with type safety and no JavaScript runtime.

This guide maps Vue concepts to rfw so you can get productive quickly.

---

## Why Go + WASM?

Vue runs in JavaScript. rfw runs in Go, compiles to WASM, and ships no JS framework bundle. Benefits:

- **Single language**, server and client are both Go. Share types, validators, domain logic.
- **Type safety**, the compiler catches mismatched signals, missing handlers, wrong prop types at build time, not at runtime.
- **SSC out of the box**, Server-Side Computed rendering with WebSocket hydration is required and built-in. No Nuxt configuration needed.
- **Fine-grained reactivity**, rfw patches only the DOM nodes that depend on changed signals, no virtual DOM diffing.

Trade-offs: rfw's ecosystem is smaller, the WASM initial load is larger than a minimal Vue bundle, and you lose access to npm packages on the client side.

---

## Mindset Shift

### Vue: Proxy-based reactivity

```js
const state = reactive({ count: 0 })
```

Vue wraps objects in JavaScript Proxies. Any property access or mutation is intercepted automatically. The system is flexible but relies on runtime tracking.

### rfw: Explicit signals with type-based detection

```go
type Counter struct {
    composition.Component
    Count *t.Int
}
```

Signals are explicit. You declare them as typed fields (`*t.Int`, `*t.String`, etc.), and `composition.New` wires them automatically based on their type. There are no hidden proxies, every reactive field is visible in the struct.

Reading and writing are explicit:

```go
c.Count.Get()  // read
c.Count.Set(5) // write
```

This is more verbose than `state.count = 5`, but every reactive access is traceable at compile time.

---

## Component Model

### Vue SFC

```vue
<template>
  <p>{{ count }}</p>
  <button @click="increment">+1</button>
</template>

<script setup>
import { ref } from 'vue'
const count = ref(0)
const increment = () => count.value++
</script>

<style scoped>
p { color: blue; }
</style>
```

### rfw struct + RTML template

```go
//go:build js && wasm

package components

import (
    "github.com/rfwlab/rfw/v2/composition"
    t "github.com/rfwlab/rfw/v2/types"
)

type Counter struct {
    composition.Component
    Count *t.Int
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}
```

`Counter.rtml` (auto-discovered from registered `embed.FS`):

```rtml
<root>
  <p>@signal:Count</p>
  <button @on:click:Increment>+1</button>
</root>
```

Key differences:

- The struct **is** the component. No `<script setup>`, no `defineProps`, no `emit`.
- The template is a separate `.rtml` file, found by convention (struct name → filename).
- State lives on the struct as signal fields. No `ref()` or `reactive()`.
- Methods on the struct are handlers. No `@click="handler"` string→function lookup ambiguity.

---

## Template Syntax Comparison

### `v-if` → `@if:`

**Vue:**

```html
<p v-if="count > 0">Positive</p>
<p v-else-if="count === 0">Zero</p>
<p v-else>Negative</p>
```

**rfw:**

```rtml
@if:Count.Get > 0
  <p>Positive</p>
@else-if:Count.Get == 0
  <p>Zero</p>
@else
  <p>Negative</p>
@endif
```

rfw uses `@if:/@else-if:/@else/@endif` blocks instead of directives on elements. Expressions reference Go struct fields or signal `.Get()`.

### `v-for` → `@for:`

**Vue:**

```html
<li v-for="item in items" :key="item.id">{{ item.text }}</li>
```

**rfw:**

```rtml
@for:item in items
  <li [key {{item.ID}}]>{{item.Text}}</li>
@endfor
```

- No `v-for` on the element itself, `@for` is a block directive.
- Keys use `[key {{expr}}]` constructors, not `:key`.
- Range syntax: `@for:i in 0..N.Get`.

### `v-model` → `@signal:...:w`

**Vue:**

```html
<input v-model="name" />
<textarea v-model="bio"></textarea>
<input type="checkbox" v-model="done" />
```

**rfw:**

```rtml
<input value="@signal:Name:w">
<textarea>@signal:Bio:w</textarea>
<input type="checkbox" checked="@signal:Done:w">
```

Append `:w` for two-way binding. Without `:w`, the binding is read-only.

### `:class` and computed bindings → `@expr:`

**Vue:**

```html
<div :class="{ active: isActive, 'text-bold': isBold }">
  {{ fullName }}
</div>
```

**Vue computed:**

```js
const fullName = computed(() => firstName.value + ' ' + lastName.value)
```

**rfw:**

```rtml
<div class="@expr:isActive && 'active' @expr:isBold && 'text-bold'">
  @expr:FirstName.Get + ' ' + LastName.Get
</div>
```

For complex logic, use a Go method:

```go
func (u *UserProfile) FullName() string {
    return u.FirstName.Get() + " " + u.LastName.Get()
}
```

```rtml
<p>{{FullName}}</p>
```

### `v-on:click` → `@on:click:`

**Vue:**

```html
<button @click="increment">+1</button>
<form @submit.prevent="save">...</form>
```

**rfw:**

```rtml
<button @on:click:Increment>+1</button>
<form @on:submit.prevent:Save>...</form>
```

Event modifiers map directly:

| Vue | rfw |
|-----|-----|
| `.prevent` | `.prevent` |
| `.stop` | `.stop` |
| `.once` | `.once` |

### `computed` → `@expr:` or Go methods

**Vue:**

```js
const doubled = computed(() => count.value * 2)
```

**rfw, inline:**

```rtml
<p>Doubled: @expr:Count.Get * 2</p>
```

**rfw, Go method:**

```go
func (c *Counter) Doubled() int {
    return c.Count.Get() * 2
}
```

```rtml
<p>Doubled: {{Doubled}}</p>
```

### `props` → `t.Prop[T]` or signal fields

**Vue:**

```js
defineProps({ title: String, count: Number })
```

**rfw, via `t.Prop[T]` fields:**

```go
type Card struct {
    composition.Component
    Title t.Prop[string]
    Count t.Prop[int]
}
```

Or via signal fields passed at construction:

```go
type Card struct {
    composition.Component
    Title *t.String
    Count *t.Int
}
```

Parent passes props when creating the component:

```go
card, err := composition.New(&Card{
    Title: t.NewString("Hello"),
    Count: t.NewInt(0),
})
if err != nil {
    log.Fatal(err)
}
```

For cross-component prop flow, use `t.Prop[T]` fields or `*t.View` for layout composition (see below).

### `emit` → handler methods

**Vue:**

```js
const emit = defineEmits(['update', 'delete'])
emit('update', newValue)
```

**rfw:** Components don't emit events. Instead, pass callbacks via dependency injection or call parent methods directly:

```go
parent.On("childUpdate", func() { /* ... */ })
```

Or use stores to communicate between components without direct coupling.

### Slots → `*t.View` field / `@include:`

**Vue parent:**

```html
<Layout>
  <template #content>Page content here</template>
</Layout>
```

**Vue Layout:**

```html
<slot name="content">Default content</slot>
```

**rfw, Layout struct:**

```go
type Layout struct {
    composition.Component
    Content *t.View
}
```

The slot name is derived from the lowercase field name (`Content` → `content`).

**Layout.rtml:**

```rtml
<root>
  <nav>My App</nav>
  <main>@include:content</main>
</root>
```

**rfw, using the layout:**

```go
layout, err := composition.New(&Layout{})
if err != nil {
    log.Fatal(err)
}
// Wire a child into the "content" slot
layout.AddDependency("content", pageView)
```

---

## State Management

### Vuex / Pinia → rfw stores

**Pinia:**

```js
export const useCounterStore = defineStore('counter', {
  state: () => ({ count: 0, name: 'rfw' }),
  getters: {
    doubled: (state) => state.count * 2,
  },
  actions: {
    increment() { this.count++ },
  },
})
```

**rfw, Store:**

```go
import "github.com/rfwlab/rfw/v2/state"

s := state.NewStore("counter", state.WithModule("app"))
s.Set("count", 0)
s.Set("name", "rfw")

// Computed
state.Map(s, "doubled", "count", func(v int) int { return v * 2 })

// Actions are just functions that call Set
func increment(s *state.Store) {
    c := s.Get("count").(int)
    s.Set("count", c+1)
}
```

**Template access:**

```rtml
<p>Count: @store:app.counter.count</p>
<p>Doubled: @store:app.counter.doubled</p>
<input value="@store:app.counter.name:w">
```

### Local state: `ref` / `reactive` → signals

**Vue:**

```js
const count = ref(0)
const user = reactive({ name: '', age: 0 })
```

**rfw:**

```go
type MyComp struct {
    composition.Component
    Count *t.Int
    Name  *t.String
    Age   *t.Int
}
```

Signals are always single values. There is no equivalent to Vue's `reactive()` for plain objects, use a `Store` for multi-key state, or separate signal fields.

---

## Routing

### Vue Router → router.Page() / router.Group()

**Vue:**

```js
const routes = [
  { path: '/', component: Home },
  { path: '/about', component: About },
  { path: '/users/:id', component: UserProfile },
]
const router = createRouter({ routes, history: createWebHistory() })
```

**rfw:**

```go
func main() {
    router.Page("/", func() *composition.View {
        v, _ := composition.New(&Home{})
        return v
    })
    router.Page("/about", func() *composition.View {
        v, _ := composition.New(&About{})
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

### Route params

**Vue:**

```js
const route = useRoute()
console.log(route.params.id)
```

**rfw:**

```go
func (u *UserProfile) OnMount() {
    params := u.HTMLComponent.RouteParams()
    id := params["id"]
}
```

### Navigation guards

**Vue:**

```js
router.beforeEach((to, from, next) => {
  if (!isAuthenticated()) next('/login')
  else next()
})
```

**rfw:**

```go
func requireAuth(params map[string]string) bool {
    return session.IsAuthenticated()
}

router.Page("/dashboard", dashboardView, requireAuth)
```

Guards are per-route functions. If any returns `false`, navigation is blocked.

---

## Lifecycle Hooks

| Vue | rfw | Notes |
|-----|-----|-------|
| `onMounted()` | `func (c *T) OnMount()` | Auto-discovered by `composition.New` |
| `onUnmounted()` | `func (c *T) OnUnmount()` | Auto-discovered |
| `onUpdated()` | _(signals auto-update DOM)_ | No explicit hook needed |
| `watch()` | `state.Effect()` or store `OnChange` | See Signals & Effects |
| `watchEffect()` | `state.Effect()` | Auto-tracks signal dependencies |

**Vue:**

```js
onMounted(() => {
  console.log('mounted')
  const timer = setInterval(() => count.value++, 1000)
  onUnmounted(() => clearInterval(timer))
})
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

---

## SSR: Nuxt → rfw SSC

SSC is **required** in rfw v2. There is no SPA fallback mode.

**Nuxt (Vue):**

1. Configure SSR in `nuxt.config.ts`
2. Server renders HTML on each request
3. Client hydrates with Vue runtime
4. Client-side navigation takes over

**rfw SSC:**

1. `rfw build` produces a WASM client bundle and a host binary
2. Host renders HTML and sends it to the browser
3. Browser loads WASM, hydrates the rendered HTML
4. A persistent WebSocket keeps client and server state synchronized
5. `h:` bindings and commands carry server-side data

```go
// Host (server side)
host.Register(host.NewHostComponent("GreetingHost", func(_ map[string]any) any {
    return map[string]any{"hostMsg": "hello from server"}
}))

sscSrv := ssc.NewSSCServer(":8080", "client")
sscSrv.ListenAndServe()
```

```rtml
<root>
  <p>Host: {h:hostMsg}</p>
  <button @on:click:h:updateTime>refresh</button>
</root>
```

Key differences from Nuxt:

- No `asyncData`, `fetch`, or `useAsync` hooks, the host component provides data directly.
- No client-side routing fallback. SSC is mandatory.
- Server-side code is Go, not JavaScript. You share types and business logic across server and client.

---

## Common Patterns Side-by-Side

### Todo List

**Vue:**

```vue
<template>
  <input v-model="newTodo" @keyup.enter="addTodo" />
  <ul>
    <li v-for="todo in todos" :key="todo.id">
      {{ todo.text }}
      <button @click="removeTodo(todo.id)">×</button>
    </li>
  </ul>
</template>

<script setup>
import { ref } from 'vue'
const newTodo = ref('')
const todos = ref([])
const addTodo = () => {
  todos.value.push({ id: Date.now(), text: newTodo.value })
  newTodo.value = ''
}
const removeTodo = (id) => {
  todos.value = todos.value.filter(t => t.id !== id)
}
</script>
```

**rfw:**

```go
type TodoApp struct {
    composition.Component
    NewTodo *t.String
    Todos   *t.Any
}

func (a *TodoApp) AddTodo() {
    todos := a.Todos.Get().([]TodoItem)
    todos = append(todos, TodoItem{ID: len(todos), Text: a.NewTodo.Get()})
    a.Todos.Set(todos)
    a.NewTodo.Set("")
}

func (a *TodoApp) OnMount() {
    a.Todos.Set([]TodoItem{})
}
```

```rtml
<root>
  <input value="@signal:NewTodo:w" @on:keyup.enter:AddTodo>
  <ul>
    @for:todo in Todos
      <li [key {{todo.ID}}]>
        {{todo.Text}}
        <button @on:click:RemoveTodo-{{todo.ID}}>×</button>
      </li>
    @endfor>
  </ul>
</root>
```

### Form with Validation

**Vue:**

```vue
<template>
  <form @submit.prevent="submit">
    <input v-model="email" />
    <span v-if="!valid">{{ error }}</span>
    <button type="submit">Submit</button>
  </form>
</template>

<script setup>
import { ref, computed } from 'vue'
const email = ref('')
const valid = computed(() => email.value.includes('@'))
const error = computed(() => valid.value ? '' : 'Invalid email')
const submit = () => { if (valid.value) { /* ... */ } }
</script>
```

**rfw:**

```go
type Form struct {
    composition.Component
    Email *t.String
}

func (f *Form) Valid() bool {
    return strings.Contains(f.Email.Get(), "@")
}

func (f *Form) Error() string {
    if f.Valid() {
        return ""
    }
    return "Invalid email"
}

func (f *Form) Submit() {
    if f.Valid() {
        // submit
    }
}
```

```rtml
<root>
  <form @on:submit.prevent:Submit>
    <input value="@signal:Email:w">
    @if:!Valid
      <span>@expr:Error</span>
    @endif
    <button type="submit">Submit</button>
  </form>
</root>
```

---

## Quick Reference: Vue → rfw

| Vue | rfw | Notes |
|-----|-----|-------|
| `ref()` | `*t.Int` / `*t.String` etc (type-detected) | Typed signals |
| `reactive()` | `*t.Store` or multiple signals | No proxy-based objects |
| `computed()` | Go method or `@expr:` | Methods auto-invoked in templates |
| `v-model` | `@signal:Name:w` | Two-way via `:w` suffix |
| `v-if` / `v-else-if` / `v-else` | `@if:` / `@else-if:` / `@else` / `@endif` | Block directives |
| `v-for` | `@for:item in items` / `@endfor` | Block directive |
| `@click` / `v-on:click` | `@on:click:Handler` | Handler is struct method |
| `:class` | `@expr:condition && 'class'` | Computed expression |
| `:style` | `style="{{expr}}"` | Expression in attribute |
| `defineProps()` | `t.Prop[T]` fields or signal fields | Auto-wired by `composition.New` |
| `defineEmits()` | Handler methods / stores | No formal emit system |
| `<slot name="x">` | `@slot:x` / `@endslot` | Slot in layout template |
| `<Comp #x="child">` | `@include:Comp` + `*t.View` field | Slot name from field name |
| `onMounted()` | `func (c *T) OnMount()` | Auto-discovered |
| `onUnmounted()` | `func (c *T) OnUnmount()` | Auto-discovered |
| `watch()` | `state.Effect()` or `store.OnChange` | |
| `provide/inject` | `*t.Inject[T]` or `Provide`/`Inject` | DI container or component tree |
| `Pinia store` | `state.NewStore()` / `*t.Store` | Namespaced key-value |
| `vue-router` | `router.Page()` / `router.Group()` | |
| `beforeEach` guard | `func(map[string]string) bool` | Per-route guards |
| `Nuxt SSR` | rfw SSC (required) | Server renders, WASM hydrates |
| `.vue` SFC | `.go` struct + `.rtml` template | Two files per component |
| `<script setup>` | Go struct definition | No separate setup function |
| `<style scoped>` | External CSS / Tailwind | No built-in scoped CSS |