//go:build js && wasm

package core

import "testing"

// A component may be linked to several host components (one per host field on
// a composition struct); registering a second name must not overwrite the
// first, and duplicates collapse.
func TestAddHostComponentKeepsAllNames(t *testing.T) {
	c := NewHTMLComponent("MultiHost", []byte(`<root></root>`), nil)
	c.AddHostComponent("Counter")
	c.AddHostComponent("Clock")
	c.AddHostComponent("Counter")

	names := c.hostComponentNames()
	if len(names) != 2 || names[0] != "Counter" || names[1] != "Clock" {
		t.Fatalf("unexpected host component names: %v", names)
	}
	if c.HostComponent != "Counter" {
		t.Fatalf("primary host component overwritten: %s", c.HostComponent)
	}
}

// Directly assigning the exported HostComponent field keeps working.
func TestHostComponentFieldFallback(t *testing.T) {
	c := NewHTMLComponent("FieldHost", []byte(`<root></root>`), nil)
	c.HostComponent = "Legacy"
	names := c.hostComponentNames()
	if len(names) != 1 || names[0] != "Legacy" {
		t.Fatalf("unexpected fallback names: %v", names)
	}
}
