//go:build js && wasm

package composition

import (
	"testing"

	core "github.com/rfwlab/rfw/v1/core"
)

func TestWrap(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	if c.HTMLComponent != hc {
		t.Fatalf("expected %v, got %v", hc, c.HTMLComponent)
	}
}

func TestUnwrap(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	if c.Unwrap() != hc {
		t.Fatalf("expected %v, got %v", hc, c.Unwrap())
	}
}

func TestWrapNilPanics(t *testing.T) {
	assertPanics(t, func() { Wrap(nil) })
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
