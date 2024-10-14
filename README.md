# rfw

rfw (Reactive Framework) is a Go-based reactive framework for building web applications using WebAssembly, 
with future plans to support native applications and the use of GL libraries.

> This is currently an experimental project, the source code is nothing more than a kind-of-working mockup.

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
<p>State is currently: @store.default.sharedState</p>
</div>
</root>
```

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

## Usage

Install the `rfw-cli`:

```bash
go install github.com/rfwlab/rfw-cli@latest
```

Create your project (currently it will create a limited example, read the code of the framework for a more complex example=:

```bash
rfw-cli init github.com/username/project-name
```

make your changes and serve it with:

```bash
rfw-cli dev
```

