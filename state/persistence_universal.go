//go:build !js || !wasm

package state

func loadPersistedState(string) map[string]any { return nil }
func saveState(string, map[string]any)         {}
