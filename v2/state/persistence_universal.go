//go:build !js || !wasm

package state

func loadPersistedState(key string) map[string]any { return nil }
func saveState(key string, state map[string]any) {}
