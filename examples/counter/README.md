# Counter

The smallest useful rfw app: one component, one reactive store, one
`@on:click` handler. No JavaScript is written.

- `components/counter_component.go` creates the `app.counter` store and
  registers the `increment` handler with `dom.RegisterHandlerFunc`.
- `components/templates/counter_component.rtml` binds the value with
  `@store:app.counter.count` and wires the button with
  `@on:click:increment`. Every `counter.Set` re-renders the binding.

## Run

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
cd examples/counter
rfw dev
```

Open http://localhost:8080 and click the button.

## Build

```bash
rfw build
```

The static bundle lands in `build/client/`.

## Compile check only

```bash
GOOS=js GOARCH=wasm go build ./...
```
