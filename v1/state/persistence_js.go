//go:build js && wasm

package state

import (
	"encoding/json"

	js "github.com/rfwlab/rfw/v1/js"
)

// loadPersistedState retrieves persisted state from localStorage.
func loadPersistedState(key string) map[string]any {
	ls := js.LocalStorage()
	if !ls.Truthy() {
		return nil
	}
	item := ls.Call("getItem", key)
	if item.Type() != js.TypeString {
		return nil
	}
	var state map[string]any
	if err := json.Unmarshal([]byte(item.String()), &state); err != nil {
		return nil
	}
	return state
}

// saveState persists the store state in localStorage.
func saveState(key string, state map[string]any) {
	ls := js.LocalStorage()
	if !ls.Truthy() {
		return
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	ls.Call("setItem", key, string(data))
}
