# Testing

rfw is a Go module, so the testing story is the Go one: `go test`.

## Native tests

Everything that does not touch the browser (stores, signals, host
components, the SSC server, the CLI) compiles natively and runs with:

```bash
go test ./...
```

Concurrency-sensitive packages are worth running with the race detector;
`state`, `host` and `ssc` are tested this way in this repository:

```bash
go test -race ./state/ ./host/ ./ssc/
```

## Browser (wasm) tests

Packages tagged `js && wasm` (`core`, `dom`, `events`, `router`,
`hostclient`, `composition`) run inside a real browser via
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest):

```bash
go install github.com/agnivade/wasmbrowsertest@latest
GOOS=js GOARCH=wasm go test -exec "$(go env GOPATH)/bin/wasmbrowsertest" ./...
```

The test binary is loaded in headless Chrome, so tests can create DOM
nodes, dispatch events and assert on rendered output. This is how the
framework tests itself; application components can do the same:

```go
func TestGreeting(t *testing.T) {
    c := core.NewHTMLComponent("Greeting",
        []byte(`<root><p>@prop:name</p></root>`),
        map[string]any{"name": "Ada"})
    c.Init(nil)
    if !strings.Contains(c.Render(), "Ada") {
        t.Fatal("prop not rendered")
    }
}
```

## Golden tests

`core/rtml_golden_test.go` pins the exact renderer output for every RTML
directive, escaping included. If you change substitution behavior, the
goldens fail and force the change to be explicit. The same technique works
for application templates whose markup must stay stable.

## Continuous integration

CI runs `go test ./...` on every push. Browser tests need Chrome on the
runner; install `wasmbrowsertest` and pass it via `-exec` as above.
