package docs

import "testing"

// TestSlugger ensures slug generation is deterministic and handles duplicates.
func TestSlugger(t *testing.T) {
	s := newSlugger()
	first := s.slug("Hello World!")
	if first != "hello-world" {
		t.Fatalf("expected 'hello-world', got %q", first)
	}
	second := s.slug("Hello World!")
	if second != "hello-world-1" {
		t.Fatalf("expected 'hello-world-1', got %q", second)
	}
}
