//go:build js && wasm

package js

import "testing"

func TestArray(t *testing.T) {
	arr := NewArray()
	if n := arr.Length(); n != 0 {
		t.Fatalf("expected empty array, got %d", n)
	}

	arr.Push("a")
	arr.Push("b")

	if n := arr.Length(); n != 2 {
		t.Fatalf("expected length 2, got %d", n)
	}

	if v := arr.Index(0).String(); v != "a" {
		t.Fatalf("expected first element 'a', got %s", v)
	}

	other := NewArray("c")
	merged := arr.Concat(other)
	if n := merged.Length(); n != 3 {
		t.Fatalf("expected length 3 after concat, got %d", n)
	}

	if v := merged.Index(2).String(); v != "c" {
		t.Fatalf("expected third element 'c', got %s", v)
	}
}
