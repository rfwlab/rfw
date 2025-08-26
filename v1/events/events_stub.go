//go:build !js || !wasm

// Package events provides no-op stubs for non-JS/WASM builds.
package events

// This file is intentionally empty to allow building on platforms
// where the real event system is unavailable.
