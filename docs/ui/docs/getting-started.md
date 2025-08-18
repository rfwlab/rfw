# Getting Started

rfw ships with a small CLI that scaffolds new projects and drives the
build system. This guide walks through installing the tooling, creating
an application and understanding how reactive updates work.

## Installing the CLI

The command line utility handles project creation and the WebAssembly
build pipeline. Install the latest version:

```bash
go install github.com/rfwlab/rfw-cli@latest
```

Verify the binary is available by running `rfw-cli -h`.

## Creating a project

Use the CLI to bootstrap a new module. The command below creates a new
Go module and pulls in the framework dependencies:

```bash
rfw-cli init github.com/username/hello-rfw
cd hello-rfw
```

The generated project contains a minimal component and a `main.go` entry
point. You can now start the development server.

## Running the development server

```bash
rfw-cli dev
```

The `dev` command compiles your Go code to WebAssembly and launches a
development server with live reloading. Every time a file changes the
page refreshes automatically.

## Building a tiny component

Templates in rfw are written in **RTML** – an HTML-like language with
reactive bindings. Create `counter.rtml`:

```rtml
<root>
  <button on:click="count++">Count: {count}</button>
</root>
```

And the accompanying component:

```go
package counter

import "github.com/rfwlab/rfw/v1/core"

//go:embed counter.rtml
var tpl []byte

type Counter struct {
    *core.HTMLComponent
    Count int
}

func New() *Counter {
    c := &Counter{HTMLComponent: core.NewHTMLComponent("Counter", tpl, nil)}
    c.SetComponent(c)
    c.Init(nil)
    return c
}
```

Mounting this component in your application gives you a working counter
with almost no JavaScript written by hand.

## Understanding reactive updates

State in rfw lives inside **stores**. A component binds to store values
by referencing them in its template. When a store value changes, the
framework updates every bound node automatically – no manual DOM
manipulation is required. Stores can expose computed values derived from
other keys and watchers that run arbitrary functions on change. See the
[state guide](./guide/state.md) for a deeper tour of these features.

## Next steps

Explore the [Why rfw?](./guide/features.md) page to learn about the
framework's design goals and continue through the rest of the guide for
details on routing, plugins and advanced patterns.
