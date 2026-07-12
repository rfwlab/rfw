# Build a real-time dashboard in 30 minutes

This tutorial builds a small live dashboard entirely in Go: a store holding
metrics, a component that renders them, a simulated data feed driven by a
goroutine and a `time.Ticker`, a list rendered with `@for`, and a pause button
wired with `@on:click`. No JavaScript is written at any point.

Prerequisites: Go 1.25+ and the rfw CLI. If you have neither, start with
[Getting started from Node](getting-started-from-node.md).

## 1. Scaffold the project

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
rfw init github.com/yourname/dashboard
cd dashboard
rfw dev
```

Open `http://localhost:8080`. You should see the scaffolded hello page. Keep
`rfw dev` running; it rebuilds on every change.

The scaffold you will touch:

```
components/
  app_component.go
  templates/app_component.rtml
pages/
  index.go        // -> route /
main.go
```

## 2. Create the metrics store

Stores are reactive key-value containers. Templates bind to them with
`@store:module.store.key` and update automatically on every `Set`.

Create `components/metrics.go`:

```go
//go:build js && wasm

package components

import "github.com/rfwlab/rfw/v2/state"

// metrics is registered globally as module "app", store "metrics",
// so templates reference it as @store:app.metrics.<key>.
var metrics = state.NewStore("metrics", state.WithModule("app"))

func seedMetrics() {
	metrics.Set("status", "live")
	metrics.Set("cpu", "0.0")
	metrics.Set("mem", "0")
	metrics.Set("requests", "0")
	metrics.Set("events", []any{})
}
```

Note the `events` key: `@for` iterates slices typed as `[]any` (each entry may
be a plain value or a `map[string]any`), so store the list as `[]any`.

## 3. The dashboard component

Create `components/dashboard_component.go`:

```go
//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

//go:embed templates/dashboard_component.rtml
var dashboardTpl []byte

type DashboardComponent struct {
	*core.HTMLComponent
	stop chan struct{}
}

func NewDashboardComponent() *DashboardComponent {
	seedMetrics()

	c := &DashboardComponent{
		HTMLComponent: core.NewHTMLComponent("DashboardComponent", dashboardTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("togglePause", togglePause)

	c.SetOnMount(func(*core.HTMLComponent) {
		c.stop = make(chan struct{})
		startFeed(c.stop)
	})
	c.SetOnUnmount(func(*core.HTMLComponent) {
		close(c.stop)
	})

	c.Init(nil)
	return c
}
```

This mirrors the scaffold's `AppComponent`: embed the template, wrap
`core.NewHTMLComponent`, call `SetComponent` and `Init`. The additions are the
handler registration and the mount/unmount hooks that own the feed's lifetime.

## 4. The template

Create `components/templates/dashboard_component.rtml`:

```html
<root>
  <div class="p-4">
    <h1 class="text-2xl font-bold">Live dashboard</h1>
    <p>Status: @store:app.metrics.status</p>
    <button @on:click:togglePause>Pause / resume</button>

    <div class="grid grid-cols-3 gap-4 my-4">
      <div>CPU: @store:app.metrics.cpu%</div>
      <div>Memory: @store:app.metrics.mem MiB</div>
      <div>Requests: @store:app.metrics.requests</div>
    </div>

    <h2 class="font-bold">Recent events</h2>
    <ul>
      @for:e in store:app.metrics.events
      <li>@prop:e.time - @prop:e.msg</li>
      @endfor
    </ul>
  </div>
</root>
```

Three directives do all the work:

- `@store:app.metrics.cpu` renders the value and re-renders on every `Set`.
- `@for:e in store:app.metrics.events ... @endfor` renders the list and
  re-renders when the `events` key changes; `@prop:e.time` reads a field of
  the current `map[string]any` entry.
- `@on:click:togglePause` dispatches clicks to the Go handler registered under
  that name. Events are delegated at the component root, so no per-element
  listeners exist.

## 5. The simulated feed

Create `components/feed.go`. A goroutine with a `time.Ticker` plays the role
of your real data source (a poller, a WebSocket, a message queue consumer):

```go
//go:build js && wasm

package components

import (
	"fmt"
	"math/rand/v2"
	"time"
)

var paused bool

func togglePause() {
	paused = !paused
	if paused {
		metrics.Set("status", "paused")
	} else {
		metrics.Set("status", "live")
	}
}

func startFeed(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		requests := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if paused {
					continue
				}
				cpu := 20 + rand.Float64()*60
				mem := 512 + rand.IntN(1024)
				requests += rand.IntN(50)

				metrics.Set("cpu", fmt.Sprintf("%.1f", cpu))
				metrics.Set("mem", fmt.Sprintf("%d", mem))
				metrics.Set("requests", fmt.Sprintf("%d", requests))

				events, _ := metrics.Get("events").([]any)
				entry := map[string]any{
					"time": time.Now().Format("15:04:05"),
					"msg":  fmt.Sprintf("cpu sample %.1f%%", cpu),
				}
				events = append([]any{entry}, events...)
				if len(events) > 8 {
					events = events[:8]
				}
				metrics.Set("events", events)
			}
		}
	}()
}
```

Every `metrics.Set` notifies the bindings created in step 4; the spans, the
`@for` list, and the status line update in place. There is no render call to
make and no diffing to think about.

To replace the simulation with real data, keep the loop shape and swap the
ticker branch for your source: read from a channel fed by an HTTP poller, or
push samples from an SSC host component.

## 6. Route it

Point the index page at the dashboard. Edit `pages/index.go`:

```go
//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/yourname/dashboard/components"
)

// Index renders the home page.
func Index() core.Component {
	return components.NewDashboardComponent()
}
```

`main.go` stays exactly as scaffolded: it imports the `pages` package and the
built-in pages plugin registers the `/` route for `Index` at build time.

## 7. Run it

If `rfw dev` is still running, the browser has already reloaded. Otherwise:

```bash
rfw dev
```

You should see CPU, memory, and request counters ticking every two seconds and
the event list filling up, newest first. Click the button: the status flips to
`paused` and the numbers freeze; click again and the feed resumes.

## Final listing

Files added or changed relative to `rfw init`:

`components/metrics.go`

```go
//go:build js && wasm

package components

import "github.com/rfwlab/rfw/v2/state"

var metrics = state.NewStore("metrics", state.WithModule("app"))

func seedMetrics() {
	metrics.Set("status", "live")
	metrics.Set("cpu", "0.0")
	metrics.Set("mem", "0")
	metrics.Set("requests", "0")
	metrics.Set("events", []any{})
}
```

`components/dashboard_component.go`

```go
//go:build js && wasm

package components

import (
	_ "embed"

	"github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
)

//go:embed templates/dashboard_component.rtml
var dashboardTpl []byte

type DashboardComponent struct {
	*core.HTMLComponent
	stop chan struct{}
}

func NewDashboardComponent() *DashboardComponent {
	seedMetrics()

	c := &DashboardComponent{
		HTMLComponent: core.NewHTMLComponent("DashboardComponent", dashboardTpl, nil),
	}
	c.SetComponent(c)

	dom.RegisterHandlerFunc("togglePause", togglePause)

	c.SetOnMount(func(*core.HTMLComponent) {
		c.stop = make(chan struct{})
		startFeed(c.stop)
	})
	c.SetOnUnmount(func(*core.HTMLComponent) {
		close(c.stop)
	})

	c.Init(nil)
	return c
}
```

`components/feed.go`

```go
//go:build js && wasm

package components

import (
	"fmt"
	"math/rand/v2"
	"time"
)

var paused bool

func togglePause() {
	paused = !paused
	if paused {
		metrics.Set("status", "paused")
	} else {
		metrics.Set("status", "live")
	}
}

func startFeed(stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		requests := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if paused {
					continue
				}
				cpu := 20 + rand.Float64()*60
				mem := 512 + rand.IntN(1024)
				requests += rand.IntN(50)

				metrics.Set("cpu", fmt.Sprintf("%.1f", cpu))
				metrics.Set("mem", fmt.Sprintf("%d", mem))
				metrics.Set("requests", fmt.Sprintf("%d", requests))

				events, _ := metrics.Get("events").([]any)
				entry := map[string]any{
					"time": time.Now().Format("15:04:05"),
					"msg":  fmt.Sprintf("cpu sample %.1f%%", cpu),
				}
				events = append([]any{entry}, events...)
				if len(events) > 8 {
					events = events[:8]
				}
				metrics.Set("events", events)
			}
		}
	}()
}
```

`components/templates/dashboard_component.rtml`

```html
<root>
  <div class="p-4">
    <h1 class="text-2xl font-bold">Live dashboard</h1>
    <p>Status: @store:app.metrics.status</p>
    <button @on:click:togglePause>Pause / resume</button>

    <div class="grid grid-cols-3 gap-4 my-4">
      <div>CPU: @store:app.metrics.cpu%</div>
      <div>Memory: @store:app.metrics.mem MiB</div>
      <div>Requests: @store:app.metrics.requests</div>
    </div>

    <h2 class="font-bold">Recent events</h2>
    <ul>
      @for:e in store:app.metrics.events
      <li>@prop:e.time - @prop:e.msg</li>
      @endfor
    </ul>
  </div>
</root>
```

`pages/index.go`

```go
//go:build js && wasm

package pages

import (
	"github.com/rfwlab/rfw/v2/core"
	"github.com/yourname/dashboard/components"
)

// Index renders the home page.
func Index() core.Component {
	return components.NewDashboardComponent()
}
```

## Where to go next

- [Dynamic lists and events](dynamic-lists.md) for the imperative flavor of
  the same pattern, where data arrives from an API call you control.
