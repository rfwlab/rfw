//go:build js && wasm

package composition

import (
	"testing"

	core "github.com/rfwlab/rfw/v2/core"
	"github.com/rfwlab/rfw/v2/dom"
	"github.com/rfwlab/rfw/v2/js"
	"github.com/rfwlab/rfw/v2/state"
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

func TestOnRegistersHandler(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	called := false
	c.On("onTest", func() { called = true })
	h := dom.GetHandler("onTest")
	if h.Type() != js.TypeFunction {
		t.Fatalf("expected function handler")
	}
	h.Invoke()
	if !called {
		t.Fatalf("expected handler to run")
	}
}

func TestOnInvalidArgs(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	assertPanics(t, func() { c.On("", func() {}) })
	assertPanics(t, func() { c.On("x", nil) })
}

func TestWrapNilPanics(t *testing.T) {
	assertPanics(t, func() { Wrap(nil) })
}

func TestProp(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	sig := state.NewSignal(1)
	c.Prop("count", sig)

	if c.HTMLComponent.Props["count"] != sig {
		t.Fatalf("expected prop to be stored")
	}
}

func TestPropOverwrite(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	first := state.NewSignal(1)
	second := state.NewSignal(2)
	c.Prop("count", first)
	c.Prop("count", second)

	if c.HTMLComponent.Props["count"] != second {
		t.Fatalf("expected latest signal in props")
	}
}

func TestStoreNamespaced(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = state.NewStoreManager()
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	s := c.Store("local")
	if state.GlobalStoreManager.GetStore(c.ID, "local") != s {
		t.Fatalf("expected store namespaced by component ID")
	}
}

func TestStoreEmptyNamePanics(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	assertPanics(t, func() { c.Store("") })
}

func TestHistoryRegistersHandlers(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = state.NewStoreManager()
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	s := c.Store("hist", state.WithHistory(5))
	s.Set("v", 1)
	s.Set("v", 2)

	c.History(s, "u", "r")

	dom.GetHandler("u").Invoke()
	if v := s.Get("v").(int); v != 1 {
		t.Fatalf("expected 1 after undo, got %d", v)
	}

	dom.GetHandler("r").Invoke()
	if v := s.Get("v").(int); v != 2 {
		t.Fatalf("expected 2 after redo, got %d", v)
	}
}

func TestHistoryInvalidArgs(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	s := c.Store("hist")
	assertPanics(t, func() { c.History(nil, "u", "r") })
	assertPanics(t, func() { c.History(s, "", "r") })
	assertPanics(t, func() { c.History(s, "u", "") })
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