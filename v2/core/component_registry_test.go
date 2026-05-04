package core

import "testing"

// minimalComponent returns nil; used only for registry registration.

type noopComponent struct{}

func (noopComponent) Render() string          { return "" }
func (noopComponent) Mount()                  {}
func (noopComponent) Unmount()                {}
func (noopComponent) OnMount()                {}
func (noopComponent) OnUnmount()              {}
func (noopComponent) GetName() string         { return "noop" }
func (noopComponent) GetID() string           { return "noop" }
func (noopComponent) SetSlots(map[string]any) {}

func TestRegisterComponentDuplicate(t *testing.T) {
	// reset registry
	ComponentRegistry = map[string]func() Component{}
	if err := RegisterComponent("dup", func() Component { return noopComponent{} }); err != nil {
		t.Fatalf("unexpected error registering component: %v", err)
	}
	if err := RegisterComponent("dup", func() Component { return noopComponent{} }); err == nil {
		t.Fatalf("expected error on duplicate registration")
	}
}

func TestMustRegisterComponentPanic(t *testing.T) {
	ComponentRegistry = map[string]func() Component{}
	MustRegisterComponent("dup", func() Component { return noopComponent{} })
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate registration")
		}
	}()
	MustRegisterComponent("dup", func() Component { return noopComponent{} })
}
