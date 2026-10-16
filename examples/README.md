# Examples

Self-contained rfw example apps. Each directory is an independent Go
module that requires the released framework and carries a `replace`
directive pointing at the repository root, so the examples always build
against this checkout.

| Example | What it shows |
| --- | --- |
| [notes](notes/) | HTTP CRUD with a real local Go API, stable IDs/versions, conflicts, draft retention and mount-scoped response guards. In-memory server, no production authentication. |
| [counter](counter/) | The minimal app: one component, one store, one `@on:click` handler. |
| [dynamic-list](dynamic-list/) | Runtime-rendered lists with `dom.RegisterHandlerElem` and `dom.ExpandEvents`; add and remove rows through event delegation. |
| [dashboard](dashboard/) | The flagship demo: a real-time dashboard with a simulated feed (goroutine plus ticker), `@for` over store data, pause/resume, and a second routed page. It is the app built in the [real-time dashboard tutorial](../docs/articles/guide/realtime-dashboard-tutorial.md). |

## Running an example

```bash
go install ./cmd/rfw
cd examples/<name>
rfw dev
```

Then open http://localhost:8080. Notes uses its own loopback API server instead
of `rfw dev`; follow its README. Install from the repository root first.

## Compile check without the CLI

```bash
cd examples/<name>
GOOS=js GOARCH=wasm go build ./...
```


## Browser acceptance

`bench/examples.mjs` uses the benchmark harness's existing Playwright dependency.
Build every example with the checkout CLI and build `examples/notes/server` as a
native executable first. Supply absolute paths for the built server and an
explicit result directory outside the source tree:

```sh
cd bench
RFW_NOTES_SERVER=/private/artifacts/notes-api \
RFW_EXAMPLE_RESULTS=/private/artifacts/example-browser \
node examples.mjs
```

The script serves the stock bundles and runs Chromium, Firefox and WebKit at
desktop/mobile viewports. It drives counter/list actions, pause/navigation/feed
recovery, real Notes CRUD/reload, two-tab conflicts and escaped text. Controlled
transport refusal and delayed real GET exercise drafts and unmount guards. The
successful CRUD path uses the actual Go API. Results and screenshots remain in
the chosen directory. Browser engines and the existing benchmark dependency
must already be available; this command does not install them.
