//go:build js && wasm

package composition

import (
	"testing"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/dom"
	"github.com/rfwlab/rfw/v1/js"
	"github.com/rfwlab/rfw/v1/state"
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
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = &state.StoreManager{modules: make(map[string]map[string]*state.Store)}
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	sig := state.NewSignal(1)
	c.Prop("count", sig)

	if c.HTMLComponent.Props["count"] != sig {
		t.Fatalf("expected prop to be stored")
	}

	store := state.GlobalStoreManager.GetStore(compositionModule, c.ID)
	if store == nil {
		t.Fatalf("expected composition store")
	}
	if store.Get("count") != sig {
		t.Fatalf("expected signal in store")
	}
}

func TestPropInvalidArgs(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)

	assertPanics(t, func() { c.Prop("", state.NewSignal(0)) })
	assertPanics(t, func() { c.Prop("k", nil) })
}

func TestPropOverwrite(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = &state.StoreManager{modules: make(map[string]map[string]*state.Store)}
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	first := state.NewSignal(1)
	second := state.NewSignal(2)
	c.Prop("count", first)
	c.Prop("count", second)

	if c.HTMLComponent.Props["count"] != second {
		t.Fatalf("expected latest signal in props")
	}
	store := state.GlobalStoreManager.GetStore(compositionModule, c.ID)
	if store.Get("count") != second {
		t.Fatalf("expected latest signal in store")
	}
}

func TestPropDoesNotOverwritePlainProp(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = &state.StoreManager{modules: make(map[string]map[string]*state.Store)}
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, map[string]any{"count": 5})
	c := Wrap(hc)
	sig := state.NewSignal(1)
	c.Prop("count", sig)

	if v := c.HTMLComponent.Props["count"].(int); v != 5 {
		t.Fatalf("must preserve existing plain prop")
	}
	store := state.GlobalStoreManager.GetStore(compositionModule, c.ID)
	if store.Get("count") != sig {
		t.Fatalf("signal must be in store")
	}
}

func TestFromPropRoundTrip(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = &state.StoreManager{modules: make(map[string]map[string]*state.Store)}
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, map[string]any{"count": 5})
	c := Wrap(hc)

	sig := c.FromProp[int]("count", 0)
	if sig.Get() != 5 {
		t.Fatalf("expected initial value 5, got %d", sig.Get())
	}
	// Props must remain the original plain value
	if v, ok := c.HTMLComponent.Props["count"].(int); !ok || v != 5 {
		t.Fatalf("expected Props[\"count\"] to remain plain 5")
	}
	// The signal is stored in the composition store
	store := state.GlobalStoreManager.GetStore(compositionModule, c.ID)
	if store.Get("count") != sig {
		t.Fatalf("expected signal stored in state store")
	}
	// FromProp should return the same signal on subsequent calls
	sigAgain := c.FromProp[int]("count", 0)
	if sigAgain != sig {
		t.Fatalf("expected FromProp to return the same signal instance")
	}

	// Create a new signal for a missing key
	sig2 := c.FromProp[int]("other", 7)
	if sig2.Get() != 7 {
		t.Fatalf("expected default value 7")
	}
	if v, ok := c.HTMLComponent.Props["other"].(*state.Signal[int]); !ok || v != sig2 {
		t.Fatalf("expected new signal in props")
	}
	if store.Get("other") != sig2 {
		t.Fatalf("expected new signal in store")
	}
}

func TestFromPropEmptyKeyPanics(t *testing.T) {
	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	assertPanics(t, func() { c.FromProp[int]("", 0) })
}

func TestFromPropIncompatibleTypePanics(t *testing.T) {
	hc := core.NewComponent("test", nil, map[string]any{"count": "x"})
	c := Wrap(hc)
	assertPanics(t, func() { c.FromProp[int]("count", 0) })
}

func TestStoreCleanupOnUnmount(t *testing.T) {
	oldGSM := state.GlobalStoreManager
	state.GlobalStoreManager = &state.StoreManager{modules: make(map[string]map[string]*state.Store)}
	defer func() { state.GlobalStoreManager = oldGSM }()

	hc := core.NewComponent("test", nil, nil)
	c := Wrap(hc)
	sig := state.NewSignal(1)
	c.Prop("count", sig)
	if state.GlobalStoreManager.GetStore(compositionModule, c.ID) == nil {
		t.Fatalf("store must exist")
	}
	c.Unmount()
	if state.GlobalStoreManager.GetStore(compositionModule, c.ID) != nil {
		t.Fatalf("store must be cleaned up")
	}
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
