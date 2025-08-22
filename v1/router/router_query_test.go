//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

// testComponent implements core.Component and routeParamReceiver for testing.
type testComponent struct {
	params map[string]string
}

func (c *testComponent) Render() string                     { return "" }
func (c *testComponent) Mount()                             {}
func (c *testComponent) Unmount()                           {}
func (c *testComponent) OnMount()                           {}
func (c *testComponent) OnUnmount()                         {}
func (c *testComponent) GetName() string                    { return "test" }
func (c *testComponent) GetID() string                      { return "" }
func (c *testComponent) SetSlots(map[string]any)            {}
func (c *testComponent) SetRouteParams(p map[string]string) { c.params = p }

func TestNavigateQueryParams(t *testing.T) {
	routes = nil
	currentComponent = nil
	RegisterRoute(Route{Path: "/query", Component: func() core.Component { return &testComponent{} }})
	Navigate("/query?key=value")
	tc, ok := currentComponent.(*testComponent)
	if !ok {
		t.Fatalf("expected testComponent, got %T", currentComponent)
	}
	if tc.params["key"] != "value" {
		t.Fatalf("expected query param 'key=value', got %v", tc.params)
	}
}

func TestNavigateNotFound(t *testing.T) {
	routes = nil
	currentComponent = nil
	called := false
	NotFoundCallback = func(p string) { called = true }
	Navigate("/missing")
	if !called {
		t.Fatalf("expected NotFoundCallback to be called")
	}
	NotFoundCallback = nil
}
