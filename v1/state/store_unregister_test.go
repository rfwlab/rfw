package state

import "testing"

func TestUnregisterStore(t *testing.T) {
	sm := &StoreManager{modules: make(map[string]map[string]*Store)}
	s := NewStore("test", WithModule("mod"))
	sm.RegisterStore("mod", "test", s)
	if sm.GetStore("mod", "test") == nil {
		t.Fatalf("expected store to be registered")
	}
	sm.UnregisterStore("mod", "test")
	if sm.GetStore("mod", "test") != nil {
		t.Fatalf("expected store to be unregistered")
	}
}
