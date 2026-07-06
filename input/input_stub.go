//go:build !js || !wasm

package input

// New creates a Manager without wiring event listeners.
func New() *Manager { return newManager() }
