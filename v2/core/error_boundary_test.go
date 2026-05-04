//go:build js && wasm

package core

import "testing"

type panicComponent struct{}

func (p *panicComponent) Render() string          { panic("boom") }
func (p *panicComponent) Mount()                  {}
func (p *panicComponent) Unmount()                {}
func (p *panicComponent) OnMount()                {}
func (p *panicComponent) OnUnmount()              {}
func (p *panicComponent) GetName() string         { return "panic" }
func (p *panicComponent) GetID() string           { return "panic" }
func (p *panicComponent) SetSlots(map[string]any) {}

func TestErrorBoundaryRender(t *testing.T) {
	eb := NewErrorBoundary(&panicComponent{}, "<div>fb</div>")
	html := eb.Render()
	expected := "<root data-component-id=\"panic\"><div>fb</div></root>"
	if html != expected {
		t.Fatalf("expected %s, got %s", expected, html)
	}
}
