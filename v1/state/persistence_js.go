//go:build js && wasm

package state

import (
	"encoding/json"
	jst "syscall/js"

	"github.com/rfwlab/rfw/v1/js"
)

// loadPersistedState retrieves persisted state from localStorage.
func loadPersistedState(key string) map[string]interface{} {
	ls := js.Global().Get("localStorage")
	if !ls.Truthy() {
		return nil
	}
	item := ls.Call("getItem", key)
	if item.Type() != jst.TypeString {
		return nil
	}
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(item.String()), &state); err != nil {
		return nil
	}
	return state
}

// saveState persists the store state in localStorage.
func saveState(key string, state map[string]interface{}) {
	ls := js.Global().Get("localStorage")
	if !ls.Truthy() {
		return
	}
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	ls.Call("setItem", key, string(data))
}
