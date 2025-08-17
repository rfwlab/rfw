//go:build js && wasm

package state

import (
	jst "syscall/js"
	"testing"
)

func TestExposeUpdateStoreBool(t *testing.T) {
	ExposeUpdateStore()
	jst.Global().Call("goUpdateStore", "mod", "test", "flag", true)
	store := GlobalStoreManager.GetStore("mod", "test")
	if store == nil {
		t.Fatalf("store not created")
	}
	v, ok := store.Get("flag").(bool)
	if !ok || !v {
		t.Fatalf("expected true bool, got %#v", store.Get("flag"))
	}
}
