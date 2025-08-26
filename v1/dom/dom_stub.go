//go:build !js || !wasm

// Package dom provides no-op stubs for non-JS/WASM builds.
package dom

// This file is intentionally empty to allow building on platforms
// where the real DOM implementation is unavailable.
