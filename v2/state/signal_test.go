package state

import "testing"

func TestSignalEffect(t *testing.T) {
	a := NewSignal(0)
	b := NewSignal(0)

	var runs int
	stop := Effect(func() func() {
		_ = a.Get()
		runs++
		return nil
	})
	defer stop()

	b.Set(1)
	if runs != 1 {
		t.Fatalf("effect ran on unrelated signal change")
	}
	a.Set(1)
	if runs != 2 {
		t.Fatalf("effect did not rerun on dependent signal change")
	}
}

func TestEffectCleanup(t *testing.T) {
	s := NewSignal(0)
	var cleans int
	stop := Effect(func() func() {
		_ = s.Get()
		return func() { cleans++ }
	})

	s.Set(1)
	if cleans != 1 {
		t.Fatalf("cleanup not called before rerun, got %d", cleans)
	}
	stop()
	if cleans != 2 {
		t.Fatalf("cleanup not called on stop, got %d", cleans)
	}
}
