# Migrating from Svelte to rfw

You know Svelte, reactive assignments, `$:` statements, stores, `{#if}` blocks. rfw shares Svelte's philosophy of fine-grained reactivity (no virtual DOM), but implements it in Go with explicit signals and type-detected field wiring. This guide maps Svelte concepts to rfw.

---

## Why Go + WASM?

Svelte compiles to vanilla JS at build time. rfw compiles Go to WASM. Both avoid runtime diffing, but:

- **Single language**, server and client are both Go. Share types, validators, and logic without serialization boundaries.
- **Compile-time safety**, Go's type system catches errors at build time that Svelte only catches at runtime.
- **SSC built-in**, Server-Side Computed rendering with WebSocket hydration is required and built-in. No SvelteKit configuration dance.
- **No bundler complexity**, `rfw build` produces the WASM bundle and host binary. No Rollup/Vite plugin juggling.

Trade-offs: rfw's ecosystem is much smaller than Svelte's. WASM initial load is heavier than Svelte's tiny output. You lose Svelte's magical `$:` reactivity and must be explicit about signal reads/writes.

---

## Mindset Shift

### Svelte: reactive assignments

```svelte
<script>
  let count = 0
  $: doubled = count * 2
  
  function increment() {
    count += 1
  }
</script>

<button on:click={increment}>{count}</button>
<p>{doubled}</p>
```

Svelte intercepts variable assignments at compile time. `count += 1` triggers reactivity automatically. `$:` statements are reactive declarations.

### rfw: explicit signals

```go
type Counter struct {
    composition.Component
    Count   *t.Int
    Doubled *t.Int
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
    c.Doubled.Set(c.Count.Get() * 2)
}
```

```rtml
<root>
  <button @on:click:Increment>@signal:Count</button>
  <p>@signal:Doubled</p>
</root>
```

Or use `@expr:` for inline computed values:

```rtml
<p>@expr:Count.Get * 2</p>
```

rfw requires explicit `.Get()` and `.Set()` calls. There is no compiler-level reactivity tracking. This is more verbose but makes every reactive access unambiguous.

---

## Component Model

### Svelte component

```svelte
<!-- Counter.svelte -->
<script>
  export let initial = 0
  let count = initial
  
  function increment() {
    count += 1
  }
</script>

<button on:click={increment}>{count}</button>
```

### rfw component

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

func (c *Counter) OnMount() {
    c.Count.Set(0)
}

func (c *Counter) Increment() {
    c.Count.Set(c.Count.Get() + 1)
}
```

`Counter.rtml` (auto-discovered):

```rtml
<root>
  <button @on:click:Increment>@signal:Count</button>
</root>
```

Key differences:

- Svelte puts script, template, and styles in one `.svelte` file. rfw splits into a `.go` struct and a `.rtml` template.
- Props are signal fields detected by type (`*t.Int`, `*t.String`, etc.). No `export let`.
- Methods on the struct are event handlers. No `on:click={handler}` closures, just method names.
- No local CSS scoping built in. Use external CSS or Tailwind.

---

## Template Syntax Comparison

### `{#if}` → `@if:`

**Svelte:**

```svelte
{#if count > 0}
  <p>Positive</p>
{:else if count === 0}
  <p>Zero</p>
{:else}
  <p>Negative</p>
{/if}
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

### `{#each}` → `@for:`

**Svelte:**

```svelte
{#each items as item, i (item.id)}
  <li>{item.text}</li>
{/each}
```

**rfw:**

```rtml
@for:item in Items
  <li [key {{item.ID}}]>{{item.Text}}</li>
@endfor
```

Range syntax:

```rtml
@for:i in 0..N.Get
  <span>{{i}}</span>
@endfor
```

### `on:click` → `@on:click:`

**Svelte:**

```svelte
<button on:click={increment}>+1</button>
<button on:click|once={launch}>Launch</button>
<form on:submit|preventDefault={save}>...</form>
```

**rfw:**

```rtml
<button @on:click:Increment>+1</button>
<button @on:click.once:Launch>Launch</button>
<form @on:submit.prevent:Save>...</form>
```

| Svelte | rfw |
|--------|-----|
| `on:click` | `@on:click:Handler` |
| `|preventDefault` | `.prevent` |
| `|stopPropagation` | `.stop` |
| `|once` | `.once` |

### `bind:value` → `@signal:...:w`

**Svelte:**

```svelte
<input bind:value={name} />
<textarea bind:value={bio} />
<input type="checkbox" bind:checked={done} />
```

**rfw:**

```rtml
<input value="@signal:Name:w">
<textarea>@signal:Bio:w</textarea>
<input type="checkbox" checked="@signal:Done:w">
```

Append `:w` for two-way binding. Without it, the binding is read-only.

### `{#await}` → OnMount + goroutines

**Svelte:**

```svelte
{#await fetchData()}
  <p>Loading...</p>
{:then data}
  <p>{data}</p>
{:catch error}
  <p>Error: {error.message}</p>
{/await}
```

**rfw:**

```go
type DataPage struct {
    composition.Component
    Data    *t.String
    Loading *t.Bool
    Error   *t.String
}

func (d *DataPage) OnMount() {
    d.Loading.Set(true)
    go func() {
        resp, err := http.Get("/api/data")
        if err != nil {
            d.Error.Set(err.Error())
            d.Loading.Set(false)
            return
        }
        var result Data
        json.NewDecoder(resp.Body).Decode(&result)
        d.Data.Set(result.Value)
        d.Loading.Set(false)
    }()
}
```

```rtml
@if:Loading.Get
  <p>Loading...</p>
@else
  @if:Error.Get != ""
    <p>Error: @signal:Error</p>
  @else
    <p>@signal:Data</p>
  @endif
@endif
```

rfw doesn't have a built-in `{#await}` block yet. Use signals and `OnMount` with goroutines.

---

## Reactivity

### Svelte reactive declarations → signals

**Svelte:**

```svelte
<script>
  let firstName = 'Ada'
  let lastName = 'Lovelace'
  $: fullName = `${firstName} ${lastName}`
</script>
```

**rfw, inline computed:**

```rtml
<p>@expr:FirstName.Get + ' ' + LastName.Get</p>
```

**rfw, Go method:**

```go
func (u *User) FullName() string {
    return u.FirstName.Get() + " " + u.LastName.Get()
}
```

```rtml
<p>{{FullName}}</p>
```

**rfw, Store computed:**

```go
s := state.NewStore("user", state.WithModule("app"))
s.Set("firstName", "Ada")
s.Set("lastName", "Lovelace")
state.Map2(s, "fullName", "firstName", "lastName", func(first, last string) string {
    return first + " " + last
})
```

```rtml
<p>@store:app.user.fullName</p>
```

### Svelte stores → rfw stores

**Svelte writable store:**

```js
import { writable } from 'svelte/store'
const count = writable(0)
count.set(5)
count.update(n => n + 1)
```

**rfw store:**

```go
s := state.NewStore("counter", state.WithModule("app"))
s.Set("count", 0)
s.Set("count", 5)
// Update:
current := s.Get("count").(int)
s.Set("count", current+1)
```

**Svelte derived store:**

```js
const doubled = derived(count, $count => $count * 2)
```

**rfw:**

```go
state.Map(s, "doubled", "count", func(v int) int { return v * 2 })
```

**Svelte readable store:**

```js
const time = readable(new Date(), (set) => {
  const interval = setInterval(() => set(new Date()), 1000)
  return () => clearInterval(interval)
})
```

**rfw:**

```go
type Clock struct {
    composition.Component
    Time *t.String
    done chan struct{}
}

func (c *Clock) OnMount() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for {
            select {
            case <-ticker.C:
                c.Time.Set(time.Now().Format("15:04:05"))
            case <-c.done:
                ticker.Stop()
                return
            }
        }
    }()
}

func (c *Clock) OnUnmount() {
    close(c.done)
}
```

---

## Component Composition

### Slots → `*t.View` field / `@include:`

**Svelte, Layout.svelte:**

```svelte
<nav>My App</nav>
<main>
  <slot name="content">
    <p>Default content</p>
  </slot>
</main>
```

**Svelte, Parent:**

```svelte
<Layout>
  <div slot="content">Custom content</div>
</Layout>
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
  <main>@slot:content
    <p>Default content</p>
  @endslot</main>
</root>
```

**Using the layout:**

```go
layout, err := composition.New(&Layout{})
if err != nil {
    log.Fatal(err)
}
layout.AddDependency("content", pageView)
```

---

## Lifecycle

| Svelte | rfw | Notes |
|--------|-----|-------|
| `onMount(fn)` | `func (c *T) OnMount()` | Auto-discovered |
| `onDestroy(fn)` | `func (c *T) OnUnmount()` | Auto-discovered |
| `beforeUpdate(fn)` | Not available | Signals update DOM directly |
| `afterUpdate(fn)` | Not available | Use `state.Effect()` if needed |
| `tick()` | Not needed | No batched updates |

**Svelte:**

```svelte
<script>
  import { onMount, onDestroy } from 'svelte'
  let interval
  onMount(() => {
    interval = setInterval(() => count += 1, 1000)
  })
  onDestroy(() => clearInterval(interval))
</script>
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

## Routing: SvelteKit → rfw

**SvelteKit (file-based routing):**

```
src/routes/+page.svelte      → /
src/routes/about/+page.svelte → /about
src/routes/users/[id]/+page.svelte → /users/:id
```

**rfw (explicit registration):**

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

rfw doesn't have file-based routing. Each route is explicitly registered with `router.Page()` or `router.Group()`.

**SvelteKit load function:**

```js
export async function load({ params }) {
  const user = await fetch(`/api/users/${params.id}`)
  return { user: await user.json() }
}
```

**rfw, OnMount with params:**

```go
type UserProfile struct {
    composition.Component
    UserName *t.String
}

func (u *UserProfile) OnMount() {
    params := u.HTMLComponent.RouteParams()
    id := params["id"]
    // fetch user data...
    u.UserName.Set(id)
}
```

---

## SSR: SvelteKit → rfw SSC

SvelteKit offers SSR as an option. rfw requires SSC.

| SvelteKit | rfw SSC | Notes |
|-----------|---------|-------|
| `+page.server.ts` load | Host component handler | Server data via `h:` bindings |
| SSR enabled by default | SSC required, no SPA fallback | |
| Adapter deployment | Host binary + WASM client | |
| Form actions | `@on:click:h:command` → host | WebSocket commands |
| No live sync | Persistent WebSocket for `h:` values | Real-time updates |

```go
// Host side
host.Register(host.NewHostComponent("PageHost", func(payload map[string]any) any {
    return map[string]any{"pageTitle": "Hello from server"}
}))
```

```rtml
<root>
  <h1>{h:pageTitle}</h1>
</root>
```

---

## Common Patterns Side-by-Side

### Reactive Counter

**Svelte:**

```svelte
<script>
  let count = 0
  $: doubled = count * 2
  function inc() { count += 1 }
</script>
<p>{count} × 2 = {doubled}</p>
<button on:click={inc}>+1</button>
```

**rfw:**

```go
type Counter struct {
    composition.Component
    Count *t.Int
}

func (c *Counter) Inc() {
    c.Count.Set(c.Count.Get() + 1)
}
```

```rtml
<root>
  <p>@signal:Count × 2 = @expr:Count.Get * 2</p>
  <button @on:click:Inc>+1</button>
</root>
```

### Todo List

**Svelte:**

```svelte
<script>
  let todos = []
  let newTodo = ''
  function addTodo() {
    todos = [...todos, { text: newTodo, id: Date.now() }]
    newTodo = ''
  }
  function removeTodo(id) {
    todos = todos.filter(t => t.id !== id)
  }
</script>
<input bind:value={newTodo} />
<button on:click={addTodo}>Add</button>
<ul>
  {#each todos as todo (todo.id)}
    <li>{todo.text} <button on:click={() => removeTodo(todo.id)}>×</button></li>
  {/each}
</ul>
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
  <input value="@signal:NewTodo:w" />
  <button @on:click:AddTodo>Add</button>
  <ul>
    @for:todo in Todos
      <li [key {{todo.ID}}]>{{todo.Text}}</li>
    @endfor>
  </ul>
</root>
```

### Shared Store

**Svelte:**

```js
// stores.js
import { writable } from 'svelte/store'
export const count = writable(0)
```

```svelte
<script>
  import { count } from './stores.js'
  function inc() { $count += 1 }
</script>
<button on:click={inc}>{$count}</button>
```

**rfw:**

```go
// Shared store creation
var CounterStore = state.NewStore("counter", state.WithModule("app"))
```

```go
type MyComp struct {
    composition.Component
    CountStore *t.Store
}

func (m *MyComp) Inc() {
    c := m.CountStore.Get("count").(int)
    m.CountStore.Set("count", c+1)
}
```

```rtml
<button @on:click:Inc>@store:app.counter.count</button>
```

---

## Quick Reference: Svelte → rfw

| Svelte | rfw | Notes |
|--------|-----|-------|
| `let x = 0` | `*t.Int` (type-detected signal) | Typed signals |
| `$: doubled = x * 2` | Go method or `@expr:` | Computed values |
| `x += 1` | `x.Set(x.Get() + 1)` | Explicit read/write |
| `bind:value={name}` | `@signal:Name:w` | Two-way binding |
| `on:click={fn}` | `@on:click:Fn` | Struct method name |
| `{#if cond}` | `@if:cond` / `@endif` | Block directives |
| `{:else if cond}` | `@else-if:cond` | |
| `{:else}` | `@else` | |
| `{#each items as item}` | `@for:item in Items` / `@endfor` | |
| `{#await promise}` | `OnMount` + goroutines | No built-in await block |
| `<slot>` | `@slot:name` / `@endslot` | |
| `<Comp let:item>` | `*t.View` field + `AddDependency` | Slot name from field name |
| `export let prop` | Signal field (type-detected) | Auto-wired |
| `createEventDispatcher` | Handler methods / stores | No built-in emit |
| `writable()` | `*t.Store` | Centralized key-value |
| `derived()` | `state.Map()` / `state.Map2()` | Computed from dependencies |
| `onMount()` | `func (c *T) OnMount()` | Auto-discovered |
| `onDestroy()` | `func (c *T) OnUnmount()` | Auto-discovered |
| SvelteKit routes | `router.Page()` / `router.Group()` | Explicit registration |
| SvelteKit SSR | rfw SSC (required) | Host renders, WASM hydrates |
| `$store` auto-sub | `@store:m.s.k` / `@signal:N` | Template bindings |
| `<style scoped>` | External CSS / Tailwind | No built-in scoped CSS |
| `{#each ... as item, i}` | `@for:item in Items` | No built-in index (use Go range) |