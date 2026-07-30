//go:build js && wasm

package core

import "testing"

func TestSuspenseMountState(t *testing.T) {
	s := NewSuspense(func() (string, error) { return "ready", nil }, "loading")
	if s.IsMounted() {
		t.Fatal("new Suspense should not be mounted")
	}
	s.Mount()
	if !s.IsMounted() {
		t.Fatal("Suspense should be mounted after Mount")
	}
	s.Unmount()
	if s.IsMounted() {
		t.Fatal("Suspense should not be mounted after Unmount")
	}
}
