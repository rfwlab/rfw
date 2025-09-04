<div align="center">
<img src="https://github.com/rfwlab/brandbook/blob/main/logos/full/png/light-full.png#gh-dark-mode-only" height="100">
<img src="https://github.com/rfwlab/brandbook/blob/main/logos/full/png/dark-full.png#gh-light-mode-only" height="100">
<hr />
<p>rfw (Reactive Framework) is a Go-based reactive framework for building web applications with WebAssembly. The framework source code lives in versioned packages such as <code>v1/core</code>, while an example application can be found in <code>docs/</code>.</p>
</div>

## Getting Started

```bash
# install the CLI
curl -L https://github.com/rfwlab/rfw/releases/download/continuous/rfw -o ~/.local/bin/rfw && chmod +x ~/.local/bin/rfw

# ensure ~/.local/bin is in your PATH, if not, add it
export PATH=$PATH:~/.local/bin

# bootstrap a project
rfw init github.com/username/project-name

# run the development server
rfw dev

# build for production
rfw build
```

By default the development server listens on port `8080`. Override it with
the `--port` flag or the `RFW_PORT` environment variable:

```bash
RFW_PORT=3000 rfw dev
```

Read the [documentation](./docs/articles/index.md) for a complete guide to the framework.

Documentation pages now include a right-hand table of contents generated from page headings for easier navigation.

## Server Side Computed (SSC)

SSC mode runs most application logic on the server while the browser loads a lightweight Wasm bundle to hydrate server-rendered HTML. The server and client keep state synchronized through a persistent WebSocket connection. See the [SSC guide](./docs/articles/guide/ssc.md) for more details.

## Build-level Plugins

`rfw` exposes a simple plugin system for build-time tasks. The CLI
automatically detects `PreBuild`, `Build` and `PostBuild` methods on plugins
and invokes them when present. Each plugin must still provide a file-watcher
trigger via `ShouldRebuild` and a numeric `Priority` to determine execution
order.

### Tailwind CSS

`rfw` includes a build step for [Tailwind CSS](https://tailwindcss.com/) using the official standalone CLI.
Place an `input.css` file (commonly under `static/`) containing the `@tailwind` directives in your project. During development the server watches
template, stylesheet and configuration files and emits a trimmed `tailwind.css`
containing only the classes you use, without requiring Node or a CDN.

### File-based Routing

The built-in `pages` plugin scans a `pages/` directory and automatically
registers routes based on its structure. Each Go file maps to a URL path:

```
pages/
  index.go        // -> /
  about.go        // -> /about
  posts/[id].go   // -> /posts/:id
```

Every file must expose a constructor using the PascalCase form of its path,
such as `func About() core.Component`. The plugin generates a temporary
`routes_gen.go` that calls `router.RegisterRoute` for each page during the
build. Import the generated package to execute the registrations, typically
via a blank import in your entrypoint:

```
import _ "your/module/pages"
```

A working example can be found under `docs/examples/pages`, which
contains `index.go`, `about.go` and `posts/[id].go` demonstrating the
generated routes `/`, `/about` and `/posts/:id`. The documentation site in
this repository uses the `pages` plugin for its home (`/`) and about (`/about`)
pages, while the `docs` plugin continues to power the documentation
content itself.

For more details and best practices, see the [Pages Plugin guide](./docs/articles/guide/pages-plugin.md).
