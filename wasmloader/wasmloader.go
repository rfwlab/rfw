//go:build js && wasm

// Package wasmloader bridges Go callers to the framework-owned browser loader.
//
// Deprecated: applications should load their primary bundle through the
// generated wasm_loader.js bootstrap. This package remains as a compatibility
// adapter for callers that start an additional bundle after the application is
// already running.
package wasmloader

import (
	"fmt"
	"strings"

	"github.com/rfwlab/rfw/v2/js"
)

// Options configures bundle loading and progress display.
type Options struct {
	Go         js.Value
	Color      string
	Height     string
	Blur       string
	SkipLoader bool
}

// Load asks the canonical browser loader to fetch and start a WebAssembly
// bundle. Delivery policy, compression negotiation and failure handling remain
// owned by wasm_loader.js rather than being duplicated in Go.
func Load(url string, opts Options) {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return
	}
	loader := js.Get("WasmLoader")
	if !loader.Truthy() || loader.Get("load").Type() != js.TypeFunction {
		js.Console().Call("error", fmt.Sprintf("cannot load %s: wasm_loader.js is not available", trimmed))
		return
	}

	config := js.NewDict()
	if opts.Go.Truthy() {
		config.Set("go", opts.Go)
	}
	if opts.Color != "" {
		config.Set("color", opts.Color)
	}
	if opts.Height != "" {
		config.Set("height", opts.Height)
	}
	if opts.Blur != "" {
		config.Set("blur", opts.Blur)
	}
	config.Set("skipLoader", opts.SkipLoader)
	loader.Call("load", trimmed, config.Value)
}
