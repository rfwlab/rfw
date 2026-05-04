package virtual

import "testing"

// TestVirtualListStub ensures the stub implementation is non-nil and callable.
func TestVirtualListStub(t *testing.T) {
	v := NewVirtualList("id", 10, 20, func(i int) string { return "" })
	if v == nil {
		t.Fatalf("expected VirtualList instance")
	}
	// Destroy should be a no-op and must not panic.
	v.Destroy()
}
