//go:build js && wasm

package shortcut

import "testing"

func TestBindAndTrigger(t *testing.T) {
	p := New()
	current = p
	triggered := false
	Bind("Control+K", func() { triggered = true })
	p.pressed["control"] = true
	p.pressed["k"] = true
	if fn, ok := p.bindings[p.combo()]; ok {
		fn()
	}
	if !triggered {
		t.Fatal("handler not triggered")
	}
	if _, ok := p.bindings["control+k"]; !ok {
		t.Fatalf("expected normalized key, got %v", p.bindings)
	}
}
