# Real-time dashboard

The flagship rfw demo, and the app built step by step in the
[real-time dashboard tutorial](../../docs/articles/guide/realtime-dashboard-tutorial.md).
The code here matches the tutorial's final listing.

A live dashboard entirely in Go:

- `components/metrics.go`: a reactive store (`app.metrics`) holding the
  metrics, including an `events` key typed `[]any` of `map[string]any`
  entries, which is what `@for` iterates.
- `components/feed.go`: a goroutine with a `time.Ticker` simulating the
  data source, updating the store every two seconds, plus the
  `togglePause` handler.
- `components/dashboard_component.go`: the component; mount/unmount hooks
  own the feed's lifetime, `@on:click:togglePause` wires the button.
- `components/templates/dashboard_component.rtml`: `@store:` bindings for
  the counters and `@for:e in store:app.metrics.events` for the event list.
- `pages/about.go`: a second page; the pages plugin maps it to `/about`
  at build time and the router handles the `<a href>` navigation.
  Navigating away unmounts the dashboard and stops the feed.

## Run

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
cd examples/dashboard
rfw dev
```

Open http://localhost:8080. The counters tick every two seconds and the
event list fills up, newest first. The button pauses and resumes the feed.

## Build

```bash
rfw build
```

## Compile check only

```bash
GOOS=js GOARCH=wasm go build ./...
```

Note: the `/about` route is registered by generated code (the pages
plugin writes `pages/routes_gen.go` during `rfw dev` / `rfw build` and
removes it afterwards), so a plain `go build` compiles the app without
that file.
