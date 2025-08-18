//go:build js && wasm

package state

import (
	"testing"

	js "github.com/rfwlab/rfw/v1/js"
)

func TestExposeUpdateStoreBool(t *testing.T) {
	ExposeUpdateStore()
	js.Call("goUpdateStore", "mod", "test", "flag", true)
	store := GlobalStoreManager.GetStore("mod", "test")
	if store == nil {
		t.Fatalf("store not created")
	}
	v, ok := store.Get("flag").(bool)
	if !ok || !v {
		t.Fatalf("expected true bool, got %#v", store.Get("flag"))
	}
}
