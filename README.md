<div align="center">
<img src="https://github.com/rfwlab/brandbook/blob/main/light-full.png?raw=true#gh-dark-mode-only" height="100">
<img src="https://github.com/rfwlab/brandbook/blob/main/dark-full.png?raw=true#gh-light-mode-only" height="100">
</div>

# rfw

rfw (Reactive Framework) is a Go-based reactive framework for building web applications with WebAssembly. The framework source code lives in versioned packages such as `v1/core`, while an example application can be found in `example/`.

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

Read the [documentation](./docs/index.md) for a complete guide to the framework.
