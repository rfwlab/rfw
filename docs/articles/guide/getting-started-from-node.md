# Getting started from Node

You know npm, `package.json`, and `npm run dev`. You have never installed Go.
This guide gets you from zero to a running rfw app without assuming any Go
background.

rfw requires **Go 1.25 or newer**.

## 1. Install Go

### Linux

Most distro packages lag behind. Prefer the official tarball:

```bash
curl -LO https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
```

Then add Go to your PATH (in `~/.bashrc` or `~/.zshrc`):

```bash
export PATH=$PATH:/usr/local/go/bin
```

If you prefer your package manager, check the version first: you need 1.25+.
On Arch `pacman -S go` is current; on Debian/Ubuntu the `golang` package is
often too old, use the tarball instead.

### macOS

```bash
brew install go
```

Or download the official `.pkg` installer from [go.dev/dl](https://go.dev/dl/)
and run it.

### Windows

Download the `.msi` installer from [go.dev/dl](https://go.dev/dl/) and run it.
It sets up PATH for you. Alternatively:

```powershell
winget install GoLang.Go
```

### Verify

```bash
go version
# go version go1.25.0 linux/amd64
```

## 2. The PATH gotcha

This is the step that trips up almost everyone coming from Node.

`go install` downloads, compiles, and drops binaries into `$(go env GOPATH)/bin`
(usually `~/go/bin`). That directory is **not** on your PATH by default. If you
run `go install ...` and then get `command not found`, this is why.

Fix it once, in your shell profile:

```bash
# ~/.bashrc or ~/.zshrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

On Windows, add `%USERPROFILE%\go\bin` to your user PATH in the environment
variables settings.

Open a new terminal (or `source` the profile) and you are done.

## 3. Mental model: npm to Go

| Node world | Go world | Notes |
| --- | --- | --- |
| `npm i -g <tool>` | `go install <module>@latest` | Installs a binary into `$(go env GOPATH)/bin` |
| `npm create <template>` | `rfw init <module-path>` | Scaffolds a new project |
| `npm run dev` | `rfw dev` | Dev server with rebuild on change |
| `package.json` | `go.mod` | Module name plus dependency list; managed by `go mod tidy` |
| `node_modules/` | module cache in `~/go/pkg/mod` | Global, deduplicated, never inside your project |
| the JS bundle | the wasm binary | One compiled artifact instead of a bundling pipeline |

Two things have no npm equivalent and need no replacement: there is no
`npm install` step after cloning (the toolchain fetches dependencies on build),
and there is no bundler config at all.

## 4. Install rfw and run your first app

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
rfw init github.com/yourname/hello
cd hello
rfw dev
```

Notes:

- `rfw init` takes a **module path** (like the `name` field in `package.json`,
  but globally unique by convention; a GitHub-style path is customary even if
  you never publish it). The project directory is the last segment, `hello`.
- `rfw dev` serves on port `8080` by default. Override with `--port`, the
  `RFW_PORT` environment variable, or a `port` field in `rfw.json`.

Open `http://localhost:8080` and you should see the scaffolded hello page.

## 5. Where to go next

- [Build a real-time dashboard in 30 minutes](realtime-dashboard-tutorial.md)
- [Dynamic lists and events](dynamic-lists.md)
