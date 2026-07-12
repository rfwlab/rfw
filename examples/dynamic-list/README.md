# Dynamic list

The pattern from the [Dynamic lists and events](../../docs/articles/guide/dynamic-lists.md)
guide: a list rendered at runtime, with add and remove, driven entirely by
event delegation.

How it works:

- The template only ships the container (`<tbody id="items-rows">`).
- `renderItems` builds the rows as a string and injects them with
  `dom.Query("#items-rows").SetHTML(dom.ExpandEvents(rows))`.
  `dom.ExpandEvents` rewrites the `@on:click:removeItem` directives in the
  runtime markup into the `data-on-*` attributes the delegated listeners
  resolve.
- `dom.RegisterHandlerElem("removeItem", ...)` receives the element that
  declared the attribute (even when the click lands on a child), so the
  row index travels in a plain `data-idx` attribute and one handler serves
  every row. Replacing the tbody's HTML never breaks anything because no
  listener is ever attached to a row.
- `dom.RegisterHandlerFunc("addItem", ...)` reads the input with
  `dom.Query("#item-name").Val()`, appends, and re-renders.

## Run

```bash
go install github.com/rfwlab/rfw/v2/cmd/rfw@latest
cd examples/dynamic-list
rfw dev
```

Open http://localhost:8080, add a few rows, remove them.

## Build

```bash
rfw build
```

## Compile check only

```bash
GOOS=js GOARCH=wasm go build ./...
```
