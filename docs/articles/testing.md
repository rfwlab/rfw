# Testing

Testing in **rfw** applications leverages Go’s standard testing tools. Components and logic can be validated with the same workflow you already use for other Go projects.

---

## Running the Test Suite

Run all tests in your project with:

```bash
go test -v ./...
```

This command executes tests for every package in the repository. Use the `-race` flag to detect data races when testing concurrent code:

```bash
go test -race ./...
```

Continuous integration environments (such as GitHub Actions or GitLab CI) typically run the same command on each push and pull request, so keeping the suite green locally helps ensure CI passes.

---

## Writing Component Tests

Although rfw compiles to WebAssembly for browser use, most component logic is just Go code and can be tested like any other package:

```go
func TestCounterIncrement(t *testing.T) {
    count := state.NewSignal(0)
    count.Set(count.Get() + 1)
    if got := count.Get(); got != 1 {
        t.Errorf("expected 1, got %d", got)
    }
}
```

Reactive primitives like **signals** and **stores** can be tested without a browser context.

---

## Testing with WebAssembly

For tests that depend on DOM integration, run them in a browser or headless environment (such as Chrome with `wasmbrowsertest`). This allows you to simulate user events and verify rendered output.

Example setup with [`wasmbrowsertest`](https://github.com/agnivade/wasmbrowsertest):

```bash
go install github.com/agnivade/wasmbrowsertest@latest
wasmbrowsertest ./...
```

---

## Recommendations

* Keep logic pure and testable outside the DOM whenever possible.
* Use Go’s race detector for concurrency safety.
* Automate tests in CI to catch regressions early.
* For DOM-specific behavior, use browser-based testing with WebAssembly.

Testing with rfw fits naturally into Go’s ecosystem, ensuring your UI logic remains reliable and maintainable.
