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

func TestDict(t *testing.T) {
	d := NewDict()
	d.Set("a", 1)
	d.Set("b", "two")

	if v := d.Get("a").Int(); v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}

	keys := d.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	found := map[string]bool{"a": false, "b": false}
	for _, k := range keys {
		if _, ok := found[k]; ok {
			found[k] = true
		}
	}
	for k, ok := range found {
		if !ok {
			t.Fatalf("missing key %s", k)
		}
	}
}

func TestMath(t *testing.T) {
	v := Math().Call("random").Float()
	if v < 0 || v >= 1 {
		t.Fatalf("random out of range: %f", v)
	}
}

func TestRegExp(t *testing.T) {
	re := RegExp().New("foo")
	if ok := re.Call("test", "foobar").Bool(); !ok {
		t.Fatalf("expected match")
	}
}
