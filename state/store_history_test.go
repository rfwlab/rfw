package state

import "testing"

func TestStoreUndoRedo(t *testing.T) {
	s := NewStore("hist", WithHistory(10))
	s.Set("count", 1)
	s.Set("count", 2)
	if v := s.Get("count"); v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}
	s.Undo()
	if v := s.Get("count"); v != 1 {
		t.Fatalf("expected 1 after undo, got %v", v)
	}
	s.Redo()
	if v := s.Get("count"); v != 2 {
		t.Fatalf("expected 2 after redo, got %v", v)
	}
}

func TestStoreHistoryLimit(t *testing.T) {
	s := NewStore("limit", WithHistory(2))
	s.Set("val", 1)
	s.Set("val", 2)
	s.Set("val", 3)
	s.Set("val", 4)
	// history limit 2 means only last two changes are tracked
	s.Undo() // 4 -> 3
	if v := s.Get("val"); v != 3 {
		t.Fatalf("expected 3 after first undo, got %v", v)
	}
	s.Undo() // 3 -> 2
	if v := s.Get("val"); v != 2 {
		t.Fatalf("expected 2 after second undo, got %v", v)
	}
	s.Undo() // no effect, history exhausted
	if v := s.Get("val"); v != 2 {
		t.Fatalf("expected 2 after exhausting history, got %v", v)
	}
}

func TestRedoClearedOnNewMutation(t *testing.T) {
	s := NewStore("redo", WithHistory(10))
	s.Set("a", 1)
	s.Set("a", 2)
	s.Undo()      // a ->1, future has mutation 1->2
	s.Set("a", 3) // new mutation should clear redo stack
	s.Redo()      // should do nothing
	if v := s.Get("a"); v != 3 {
		t.Fatalf("expected 3 after redo with cleared history, got %v", v)
	}
}
