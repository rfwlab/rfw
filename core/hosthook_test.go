//go:build js && wasm

package core

import "testing"

// Rendering a component that declares host components goes through the hook
// the hostclient package installs. Without hostclient linked the hook is nil
// and rendering must still succeed, which is what keeps client-only builds
// free of the websocket and net/http stacks.
func TestHostRegisterHook(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })

	type call struct {
		id, name string
		vars     []string
	}
	var calls []call
	SetHostRegister(func(id, name string, vars []string) {
		calls = append(calls, call{id: id, name: name, vars: vars})
	})

	c := NewHTMLComponent("HookHost", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.AddHostComponent("Clock")
	c.Render()

	if len(calls) != 2 {
		t.Fatalf("expected 2 host registrations, got %d: %+v", len(calls), calls)
	}
	if calls[0].name != "Counter" || calls[1].name != "Clock" {
		t.Fatalf("unexpected registration order: %+v", calls)
	}
	if calls[0].id != c.ID {
		t.Fatalf("expected component id %q, got %q", c.ID, calls[0].id)
	}
}

// A nil hook is the client-only case: rendering must not panic.
func TestHostRegisterHookNil(t *testing.T) {
	prev := hostRegisterComponent
	t.Cleanup(func() { hostRegisterComponent = prev })
	hostRegisterComponent = nil

	c := NewHTMLComponent("NoHook", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	if got := c.Render(); got == "" {
		t.Fatal("expected a rendered template with no host hook installed")
	}
}
