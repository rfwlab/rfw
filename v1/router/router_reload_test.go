//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type reloadComponent struct{}

func (c *reloadComponent) Render() string          { return "" }
func (c *reloadComponent) Mount()                  {}
func (c *reloadComponent) Unmount()                {}
func (c *reloadComponent) OnMount()                {}
func (c *reloadComponent) OnUnmount()              {}
func (c *reloadComponent) GetName() string         { return "reload" }
func (c *reloadComponent) GetID() string           { return "" }
func (c *reloadComponent) SetSlots(map[string]any) {}

func TestNavigateReloadsRouteEachTime(t *testing.T) {
	Reset()
	count := 0
	RegisterRoute(Route{Path: "/reload", Component: func() core.Component {
		count++
		return &reloadComponent{}
	}})
	Navigate("/reload")
	Navigate("/reload")
	if count != 2 {
		t.Fatalf("expected loader called twice, got %d", count)
	}
}
