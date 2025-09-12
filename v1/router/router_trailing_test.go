//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type trailingComponent struct{}

func (c *trailingComponent) Render() string          { return "" }
func (c *trailingComponent) Mount()                  {}
func (c *trailingComponent) Unmount()                {}
func (c *trailingComponent) OnMount()                {}
func (c *trailingComponent) OnUnmount()              {}
func (c *trailingComponent) GetName() string         { return "trailing" }
func (c *trailingComponent) GetID() string           { return "" }
func (c *trailingComponent) SetSlots(map[string]any) {}

func TestNavigateTrailingSlash(t *testing.T) {
	Reset()
	RegisterRoute(Route{Path: "/trail", Component: func() core.Component { return &trailingComponent{} }})
	Navigate("/trail/")
	if _, ok := currentComponent.(*trailingComponent); !ok {
		t.Fatalf("expected trailingComponent with trailing slash, got %T", currentComponent)
	}
}

func TestNavigateTrailingSlashNotFound(t *testing.T) {
	Reset()
	RegisterRoute(Route{Path: "/trail", Component: func() core.Component { return &trailingComponent{} }})
	called := false
	NotFoundCallback = func(p string) { called = true }
	Navigate("/trail/extra")
	if !called {
		t.Fatalf("expected NotFoundCallback for extra path, got none")
	}
	NotFoundCallback = nil
}
