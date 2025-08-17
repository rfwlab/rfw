<div align="center">
<img src="https://github.com/rfwlab/brandbook/blob/main/light-full.png?raw=true#gh-dark-mode-only" height="100">
<img src="https://github.com/rfwlab/brandbook/blob/main/dark-full.png?raw=true#gh-light-mode-only" height="100">

<hr />

<p>rfw (Reactive Framework) is a Go-based reactive framework for building web applications using WebAssembly, 
with future plans to support native applications and the use of GL libraries.</p>
</div>

> This is currently an experimental project, the source code is nothing more than a kind-of-working mockup.

The repository is organized so that the framework lives at the root under versioned packages (e.g. `v1/core`, `v1/router`).
An example application demonstrating the current capabilities is available in the `example/` directory.

## The idea

The idea behind rfw is to create a highly performant and reactive framework that leverages Go and WebAssembly, 
providing a simplified and native experience in web development. Unlike many other frameworks, rfw completely 
avoids the complexities of virtual DOM management and does not implement diffing or proxy/virtual dom systems, 
instead it relies on an event-driven reactive update system, taking full advantage of the native performance 
of WebAssembly to update only the parts of the DOM that change based on state.

## Reactivity Implementation

Reactivity in rfw is based on a **direct binding system** between state variables and the DOM; each reactive
variable (`ReactiveVar`) is connected to DOM elements through listeners that are registered at component
rendering level. When the value of a reactive variable changes, the framework automatically updates the
portions of the DOM associated with that variable, without recalculating or differentiating the entire DOM
structure, minimizing unnecessary updates that could impact the browser.

### Computed Values and Watchers

The state package also provides **computed values** and **watchers** for handling derived state and side
effects. A computed value derives its output from one or more store keys and is re-evaluated whenever one of
its dependencies changes. Watchers observe specific keys and execute a callback after the state updates.

```go
store := state.NewStore("profile", state.WithModule("user"), state.WithPersistence())

// register a computed value
store.RegisterComputed(state.NewComputed("fullName", []string{"first", "last"}, func(s map[string]interface{}) interface{} {
    return s["first"].(string) + " " + s["last"].(string)
}))

// react to changes
store.RegisterWatcher(state.NewWatcher([]string{"fullName"}, func(s map[string]interface{}) {
    fmt.Println("name changed to", s["fullName"])
}))

// best practice: set dependencies first so computeds are evaluated correctly
store.Set("first", "Ada")
store.Set("last", "Lovelace")
```

Use computed values to keep derived data in sync without manual bookkeeping and watchers for side effects
such as logging or triggering network calls.

State values can optionally persist to `localStorage` when a store is created with `state.WithPersistence()`. Temporary
stores omit this option and reset on reload. The example application includes a `/stores` route demonstrating both
behaviors side by side.

### Updating state from JavaScript

Call `state.ExposeUpdateStore()` to register a global `goUpdateStore(module, store, key, value)` function. It accepts
strings, numbers, booleans and other serializable values, letting external scripts change store entries with
their native types.

## Logging

rfw uses a pluggable logger exposed through the `core` package. By default log messages are printed to
`stdout`, but applications can replace the logger to integrate with custom logging solutions. Implement the
`core.Logger` interface and register it with `core.SetLogger`:

```go
type myLogger struct{}

func (myLogger) Debug(f string, v ...interface{}) { /* ... */ }
func (myLogger) Info(f string, v ...interface{})  { /* ... */ }
func (myLogger) Warn(f string, v ...interface{})  { /* ... */ }
func (myLogger) Error(f string, v ...interface{}) { /* ... */ }

func init() {
    core.SetLogger(myLogger{})
}
```

All framework packages log through this interface, allowing easy redirection to structured loggers or other
destinations.

## Components Types

Components are the primary entities used to build applications. In rfw there are 2 different type of 
components: HTMLComponent(s) and GLComponent(s).

### 1. **HTMLComponent**

The **HTMLComponent** represents standard web components that are rendered directly into the DOM using RTML 
(generated HTML); these components are ideal for creating web applications such as portals, blogs, PWAs 
and possibly mini-games based on the DOM.

#### RTML (Reactive Templating Markup Language)

**RTML** is the templating language used in rfw. It allows easy interpolation of dynamic data and lifecycle 
management of components in a reactive way, similar to frameworks like Vue or React, but built specifically 
for Go. This allows developers to create dynamic interfaces in an intuitive way.

RTML code example:

```html
<root>
  @include:header
  <div class="p-4 pt-0">
    @include:card
    <p>State is currently: @store:app.default.sharedState</p>
  </div>
</root>
```

#### Event Handling

RTML offers two ways to wire DOM events to Go code.

##### Standard registry-based bindings

Use directive-style attributes such as `@click:toggle` (or the explicit `@on:click:toggle`).
They are translated into `data-on-click` attributes and listeners are resolved via a Go
registry, so no JavaScript needs to be written.

```html
<button @click:toggle>Toggle</button>
```

Register handlers from Go:

```go
import "github.com/rfwlab/rfw/v1/dom"

func init() {
    dom.RegisterHandlerFunc("toggle", func() {
        fmt.Println("clicked")
    })
}
```

##### Legacy exposed handlers (deprecated)

For compatibility, functions can still be exposed to the JavaScript global scope and
referenced with the `@on:` prefix. This requires `jsa.Expose` and should be avoided in new
code.

```html
<button @on:click:toggle>Toggle</button>
```

```go
import jsa "github.com/rfwlab/rfw/v1/js"

func init() {
    jsa.Expose("toggle", func() {
        fmt.Println("clicked")
    })
}
```

#### DOM Events and Observers

The `v1/events` package exposes browser events and observers as Go channels. These helpers can
replace inline `onclick` handlers and other DOM events, letting you handle interactions entirely in Go.

```go
btn := js.Global().Get("document").Call("getElementById", "clickBtn")
clicks := events.Listen("click", btn)
go func() {
    for range clicks {
        // handle button clicks
    }
}()

mutCh, stopMut := events.ObserveMutations("#node")
defer stopMut()
go func() {
    for m := range mutCh {
        _ = m // process mutation records
    }
}()

intCh, stopInt := events.ObserveIntersections(".watched", js.ValueOf(map[string]any{}))
defer stopInt()
go func() {
    for entry := range intCh {
        _ = entry // process intersection entries
    }
}()
```

##### Why use channel-based events?

- Keep event logic in Go without embedding inline JavaScript or exposing global handlers.
- Stream events through Go channels, making it simple to fan-out work across goroutines.
- Each helper returns a cleanup function so listeners and observers can be released, preventing memory leaks.
- Serves as a path toward deprecating template `@on` attributes and ad‑hoc JS helpers.

See the `example/components` directory for usage demonstrations.

#### Lifecycle Hooks

Components can react when they are added to or removed from the DOM using the `OnMount` and `OnUnmount` hooks. The `HTMLComponent` provides default no-op implementations so you can override only the hooks you need. When embedding `HTMLComponent`, register the struct itself with `SetComponent` so that lifecycle hooks are invoked.

```go
type HeaderComponent struct {
    *core.HTMLComponent
}

func (c *HeaderComponent) OnMount() {
    fmt.Println("header mounted")
}

func (c *HeaderComponent) OnUnmount() {
    fmt.Println("header unmounted")
}

func NewHeaderComponent() *HeaderComponent {
    c := &HeaderComponent{HTMLComponent: core.NewHTMLComponent("HeaderComponent", tpl, nil)}
    c.SetComponent(c)
    c.Init(nil)
    return c
}
```

The example application uses these hooks to count how many times the navigation header is mounted and unmounted.

### 2. **GLComponent**

_Work in progress._

The **GLComponent** will introduce support for rendering complex graphical components using OpenGL or WebGL. 
The idea is to create a markup language that acts as an intermediary between the 2 different technologies, 
allowing to build both web-based applications (WebGL/Canvas) and native applications via OpenGL and Vulkan, 
in this last case it is rfw that draws the window. The plans include interpolation between HTMLComponent(s) 
and GLComponent(s).

Some examples of use include: Development of simple and complex games (my idea is to create a game engine as
an exercise), advanced data and graphics visualization, and development of native applications for all devices
with OpenGL and Vulkan support.

## Router

The `v1/router` package provides client-side navigation with support for nested routes, navigation guards and
lazy loaded components. Components are created only when their route becomes active.

```go
router.RegisterRoute(router.Route{
    Path: "/parent",
    Component: func() core.Component { return components.NewParentComponent() },
    Children: []router.Route{
        {
            Path: "/parent/child",
            Component: func() core.Component { return components.NewChildComponent() },
        },
    },
    Guards: []router.Guard{func(params map[string]string) bool {
        return true // block navigation by returning false
    }},
})
```

Guards run before navigation and can cancel the transition when they return `false`. See the `example/`
directory for a working demonstration. The example's header shows whether the protected route is enabled;
click the "Unlock Protected" button to update the store, then use the "Protected" link to navigate to the
guarded page. If a guard blocks the first navigation (for example by visiting `/protected` directly), the
router falls back to the root route instead of rendering a blank page.

## Usage

Install the `rfw-cli`:

```bash
go install github.com/rfwlab/rfw-cli@latest
```

Create your project (currently it will create a limited example, read the code of the framework for a more 
complex example:

```bash
rfw-cli init github.com/username/project-name
```

make your changes and serve it with:

```bash
rfw-cli dev
```

## Contributing

This project is currently in the experimental phase, so any help is welcome. If you want to contribute,
please feel free to open an issue or a pull request, we are open to any suggestions and improvements.

### My IDE is not recognizing the RTML syntax

If you are using Visual Studio Code, look for a pop-up in the bottom right corner of the screen that prompts
you to install the recommended extensions for the project.

Other IDEs are not yet supported in the current state of the project. I suggest to just map the `.rtml`
extension to HTML in your IDE settings, you will not see the syntax highlighting for the RTML specific
syntax, but it will work.

### My IDE is not recognizing imports

This happens because the project has specific build tags in most of the files (`//go:build js && wasm`).
To fix this just set the build tags in your IDE settings to `js,wasm`.

#### VSCode

Add the following to your `settings.json`:

```json
{
  "go.buildTags": "js wasm"
}
```

#### Zed IDE

Add the following to your `settings.json` (extending the `lsp` key):

```json
{
  "lsp": {
    "gopls": {
      "initialization_options": {
        "buildFlags": ["-tags=js,wasm"]
      }
    }
  }
}
```

#### Other IDEs

Please refer to the documentation of your IDE to find out how to set build tags for Go projects.

### Proper LSP support for RTML

Currently there is no LSP support for RTML, so you will not have autocompletion or syntax checking
for RTML files.
