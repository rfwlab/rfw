//go:build js && wasm

package dom

// Binding precompilation is inert. The parsed result was only ever written to
// a map nothing read, and event wiring is done by core's own template pass in
// core/rtml.go. The two functions below stay so their callers keep compiling,
// but they no longer parse the template: doing so pulled golang.org/x/net/html,
// a full server-side HTML parser, into every browser bundle.

// RegisterBindings associates bindings with a component instance. It is a
// no-op, kept for callers that still invoke it.
func RegisterBindings(string, string, string) {}

// OverrideBindings replaces the cached bindings for a component name. It is a
// no-op, kept for callers that still invoke it.
func OverrideBindings(string, string) {}
