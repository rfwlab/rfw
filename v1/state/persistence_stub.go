//go:build !js || !wasm

package state

// loadPersistedState is a no-op on non-JS platforms.
func loadPersistedState(key string) map[string]any { return nil }

// saveState is a no-op on non-JS platforms.
func saveState(key string, state map[string]any) {}
