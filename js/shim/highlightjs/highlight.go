//go:build js && wasm

// Package highlightjs connects Go language definitions to Highlight.js.
package highlightjs

import (
	js "github.com/rfwlab/rfw/v2/js"
)

// RegisterLanguage wires a Highlight.js language definition from Go.
// The def callback receives the hljs object and must return the language config.
func RegisterLanguage(name string, def func(hljs js.Value) js.Value) {
	hljsObj := js.Get("hljs")
	hljsObj.Call("registerLanguage", name, js.SafeFuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			return def(args[0])
		}
		return nil
	}))
}
