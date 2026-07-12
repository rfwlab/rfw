# Examples

Self-contained rfw example apps. Each directory is an independent Go
module that requires the released framework and carries a `replace`
directive pointing at the repository root, so the examples always build
against this checkout.

| Example | What it shows |
| --- | --- |
| [counter](counter/) | The minimal app: one component, one store, one `@on:click` handler. |
| [dynamic-list](dynamic-list/) | Runtime-rendered lists with `dom.RegisterHandlerElem` and `dom.ExpandEvents`; add and remove rows through event delegation. |
| [dashboard](dashboard/) | The flagship demo: a real-time dashboard with a simulated feed (goroutine plus ticker), `@for` over store data, pause/resume, and a second routed page. It is the app built in the [real-time dashboard tutorial](../docs/articles/guide/realtime-dashboard-tutorial.md). |

## Running an example

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
cd examples/<name>
rfw dev
```

Then open http://localhost:8080.

## Compile check without the CLI

```bash
cd examples/<name>
GOOS=js GOARCH=wasm go build ./...
```
