//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type testComponent struct{}

func (c *testComponent) Render() string          { return "" }
func (c *testComponent) Mount()                  {}
func (c *testComponent) Unmount()                {}
func (c *testComponent) OnMount()                {}
func (c *testComponent) OnUnmount()              {}
func (c *testComponent) GetName() string         { return "test" }
func (c *testComponent) GetID() string           { return "" }
func (c *testComponent) SetSlots(map[string]any) {}

func TestCanNavigate(t *testing.T) {
	Reset()
	RegisterRoute(Route{Path: "/can", Component: func() core.Component { return &testComponent{} }})
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
