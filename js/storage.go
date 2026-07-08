//go:build js && wasm

package js

// StorageGet reads a key from localStorage, returning "" when absent.
func StorageGet(key string) string {
	v := LocalStorage().Call("getItem", key)
	if v.Type() == TypeString {
		return v.String()
	}
	return ""
}

// StorageSet writes a key/value pair to localStorage.
func StorageSet(key, value string) { LocalStorage().Call("setItem", key, value) }

// StorageRemove deletes a key from localStorage.
func StorageRemove(key string) { LocalStorage().Call("removeItem", key) }
