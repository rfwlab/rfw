//go:build devtools

package state

import "testing"

func TestSignalHookAndSnapshot(t *testing.T) {
	var gotID int
	var gotVal any
	SignalHook = func(id int, v any) {
		gotID = id
		gotVal = v
	}
	s := NewSignal(1)
	s.Set(2)
	if gotID != s.id || gotVal != 2 {
		t.Fatalf("hook not invoked")
	}
	snap := SnapshotSignals()
	if v, ok := snap[s.id]; !ok || v != 2 {
		t.Fatalf("snapshot missing or incorrect, got %v", snap)
	}
}
