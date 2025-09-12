# js

Thin wrappers around `syscall/js` for interacting with the browser.
The package re-exports `syscall/js` types and exposes helpers for
common globals.

| Function | Description |
| --- | --- |
| `Global()` | Returns the JavaScript global object. |
| `Window()` | Shortcut for `window`. |
| `Document()` | Shortcut for `document`. |
| `Call(name, args...)` | Invokes a function on the global object. |
| `ValueOf(v)` | Converts a Go value to a JavaScript value. |
| `TypedArrayOf(slice)` | Converts a Go slice to a JavaScript typed array. Call `Release` on the returned value when done. |
| `FuncOf(fn)` | Wraps a Go function for use in JavaScript. |
| `Expose(name, fn)` | Registers a no-arg Go function on the global scope. |
| `ExposeEvent(name, fn)` | Registers a Go function that receives the first argument as an event. |
| `ExposeFunc(name, fn)` | Registers a Go function with custom arguments on the global scope. |
| `RequestAnimationFrame(cb)` | Wrapper for `requestAnimationFrame`. |
| `Fetch(args...)` | Wrapper for the global `fetch` function. |

Additional helpers provide access to common objects like `Console()`,
`History()`, `LocalStorage()` and constructors such as
`MutationObserver()` and `IntersectionObserver()`.

For higher-level HTTP helpers built on top of `fetch`, see the
[`http` package](./http).

## Usage

Use the `js` package for all direct JavaScript interop:

```go
import js "github.com/rfwlab/rfw/v1/js"

js.Expose("goHello", func() {
        js.Console().Call("log", "hello from Go")
})
```

Direct imports of `syscall/js` should be limited to type references.
