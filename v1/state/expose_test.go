//go:build js && wasm

package state

import (
	"syscall/js"
	"testing"
)

func TestExposeUpdateStoreBool(t *testing.T) {
	ExposeUpdateStore()
	js.Global().Call("goUpdateStore", "test", "flag", true)
	store := GlobalStoreManager.GetStore("test")
	if store == nil {
		t.Fatalf("store not created")
	}
	v, ok := store.Get("flag").(bool)
	if !ok || !v {
		t.Fatalf("expected true bool, got %#v", store.Get("flag"))
	}
}
