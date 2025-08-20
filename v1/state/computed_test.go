package state

import (
	"strings"
	"testing"
)

func TestComputedStability(t *testing.T) {
	s := NewStore("test")
	s.Set("a", 1)

	evalCount := 0
	c := NewComputed("double", []string{"a"}, func(m map[string]any) any {
		evalCount++
		return m["a"].(int) * 2
	})
	s.RegisterComputed(c)

	if evalCount != 1 {
		t.Fatalf("expected 1 evaluation, got %d", evalCount)
	}
	if v := s.Get("double"); v != 2 {
		t.Fatalf("expected computed value 2, got %v", v)
	}

	// Setting dependency to same value should not re-evaluate
	s.Set("a", 1)
	if evalCount != 1 {
		t.Fatalf("computed re-evaluated without dependency change")
	}

	// Setting unrelated key should not re-evaluate
	s.Set("b", 5)
	if evalCount != 1 {
		t.Fatalf("computed re-evaluated for unrelated key")
	}

	// Changing dependency should trigger re-evaluation
	s.Set("a", 3)
	if evalCount != 2 {
		t.Fatalf("expected second evaluation after dependency change, got %d", evalCount)
	}
	if v := s.Get("double"); v != 6 {
		t.Fatalf("expected computed value 6, got %v", v)
	}
}

func TestMapHelpers(t *testing.T) {
	s := NewStore("test")
	s.Set("count", 2)

	Map(s, "double", "count", func(v int) int { return v * 2 })

	if v := s.Get("double"); v != 4 {
		t.Fatalf("expected 4, got %v", v)
	}

	s.Set("count", 3)
	if v := s.Get("double"); v != 6 {
		t.Fatalf("expected 6 after update, got %v", v)
	}

	s.Set("first", "Ada")
	s.Set("last", "Lovelace")

	Map2(s, "fullName", "first", "last", func(f, l string) string {
		return strings.TrimSpace(f + " " + l)
	})

	if v := s.Get("fullName"); v != "Ada Lovelace" {
		t.Fatalf("expected full name Ada Lovelace, got %v", v)
	}

	s.Set("last", "Hopper")
	if v := s.Get("fullName"); v != "Ada Hopper" {
		t.Fatalf("expected full name Ada Hopper, got %v", v)
	}
}
