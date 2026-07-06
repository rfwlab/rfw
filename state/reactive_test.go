package state

import (
	"reflect"
	"testing"
)

func TestReactiveVarInt(t *testing.T) {
	rv := NewReactiveVar(0)
	var changed int
	rv.OnChange(func(v int) { changed = v })
	rv.Set(42)
	if got := rv.Get(); got != 42 {
		t.Fatalf("expected Get to return 42, got %d", got)
	}
	if changed != 42 {
		t.Fatalf("expected OnChange to fire with 42, got %d", changed)
	}
}

type sample struct {
	A int
	B string
}

func TestReactiveVarStruct(t *testing.T) {
	initial := sample{A: 1, B: "foo"}
	rv := NewReactiveVar(initial)
	var changed sample
	rv.OnChange(func(s sample) { changed = s })
	newVal := sample{A: 2, B: "bar"}
	rv.Set(newVal)
	if got := rv.Get(); !reflect.DeepEqual(got, newVal) {
		t.Fatalf("expected Get to return %v, got %v", newVal, got)
	}
	if !reflect.DeepEqual(changed, newVal) {
		t.Fatalf("expected OnChange to receive %v, got %v", newVal, changed)
	}
}
