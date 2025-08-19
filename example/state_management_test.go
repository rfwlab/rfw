package main

import (
	"testing"

	"github.com/rfwlab/rfw/v1/state"
)

func TestComputedDouble(t *testing.T) {
	s := state.NewStore("test")
	s.Set("count", 2)
	s.RegisterComputed(state.NewComputed("double", []string{"count"}, func(m map[string]interface{}) interface{} {
		return m["count"].(int) * 2
	}))
	if v := s.Get("double"); v != 4 {
		t.Fatalf("expected 4, got %v", v)
	}
	s.Set("count", 3)
	if v := s.Get("double"); v != 6 {
		t.Fatalf("expected 6 after update, got %v", v)
	}
}
