//go:build js && wasm

package main

import (
	"testing"

	core "github.com/rfwlab/rfw/v1/core"
	"github.com/rfwlab/rfw/v1/router"
)

type mountCheckComponent struct {
	mounted bool
}

func (c *mountCheckComponent) Render() string          { return "" }
func (c *mountCheckComponent) Mount()                  { c.mounted = true }
func (c *mountCheckComponent) Unmount()                {}
func (c *mountCheckComponent) OnMount()                {}
func (c *mountCheckComponent) OnUnmount()              {}
func (c *mountCheckComponent) GetName() string         { return "mount-check" }
func (c *mountCheckComponent) GetID() string           { return "" }
func (c *mountCheckComponent) SetSlots(map[string]any) {}

func TestGuardBlocksNavigation(t *testing.T) {
	router.Reset()
	mc := &mountCheckComponent{}
	router.RegisterRoute(router.Route{
		Path:      "/guarded",
		Component: func() core.Component { return mc },
		Guards: []router.Guard{
			func(params map[string]string) bool { return false },
		},
	})
	router.Navigate("/guarded")
	if mc.mounted {
		t.Fatalf("expected guard to block navigation")
	}
}
