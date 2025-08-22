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
```

Read the [documentation](./docs/articles/index.md) for a complete guide to the framework.

## Build-level Plugins

`rfw` exposes a simple plugin system for build-time tasks. Plugins can register
build steps and file-watcher triggers to extend the CLI without relying on
external tooling.

### Tailwind CSS

`rfw` includes a build step for [Tailwind CSS](https://tailwindcss.com/) using the official standalone CLI.
Place an `input.css` file containing the `@tailwind` directives in your project. During development the server watches
template, stylesheet and configuration files and emits a trimmed `tailwind.css`
containing only the classes you use, without requiring Node or a CDN.
