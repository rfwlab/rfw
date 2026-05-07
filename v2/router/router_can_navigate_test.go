//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v2/core"
)

type canNavigateComponent struct{}

func (c *canNavigateComponent) Render() string             { return "" }
func (c *canNavigateComponent) Mount()                     {}
func (c *canNavigateComponent) Unmount()                   {}
func (c *canNavigateComponent) OnMount()                   {}
func (c *canNavigateComponent) OnUnmount()                 {}
func (c *canNavigateComponent) GetName() string            { return "test" }
func (c *canNavigateComponent) GetID() string              { return "" }
func (c *canNavigateComponent) SetSlots(map[string]any)    {}
func (c *canNavigateComponent) IsMounted() bool            { return false }
func (c *canNavigateComponent) OnParams(map[string]string) {}

func TestCanNavigate(t *testing.T) {
	Reset()
	RegisterRoute(Route{Path: "/can", Component: func() core.Component { return &canNavigateComponent{} }})
	if !CanNavigate("/can") {
		t.Fatalf("expected true for registered route")
	}
	if !CanNavigate("/can?foo=bar") {
		t.Fatalf("expected true for registered route with query")
	}
	if CanNavigate("/missing") {
		t.Fatalf("expected false for unregistered route")
	}
}
