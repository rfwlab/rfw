# Requirements

Before you start building with **rfw**, make sure your environment is ready.

## Go

- Version **1.21** or later
- Installed and available on your `PATH`
- Familiarity with the Go module system (`go mod init`, `go mod tidy`)

## WebAssembly

- Go's `GOOS=js GOARCH=wasm` target must compile successfully
- The `rfw` CLI handles Wasm builds automatically, but custom tooling should use this target

## Browser

A modern browser with WebAssembly support:

- Chrome
- Firefox
- Safari
- Edge

## System Tools

- `git` for version control and fetching dependencies
- Internet connection for downloading modules and assets

## Knowledge Prerequisites

- Basic **Go** programming
- Fundamentals of **HTML** and **CSS**
- Basic **JavaScript** (to understand integration points)

## Optional

- Familiarity with concepts of **reactivity** and component-based UI
- Basic understanding of server-side Go if you plan to write host components

---

With these requirements in place, continue with [Quick Start](/docs/getting-started/quick-start) to scaffold your first project.