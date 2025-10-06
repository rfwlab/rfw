//go:build js && wasm

package router

import (
	"testing"

	"github.com/rfwlab/rfw/v1/core"
)

type routeComponent struct{}

func (routeComponent) Render() string          { return "" }
func (routeComponent) Mount()                  {}
func (routeComponent) Unmount()                {}
func (routeComponent) OnMount()                {}
func (routeComponent) OnUnmount()              {}
func (routeComponent) GetName() string         { return "route" }
func (routeComponent) GetID() string           { return "route" }
func (routeComponent) SetSlots(map[string]any) {}

func TestRegisteredRoutes(t *testing.T) {
	Reset()
	RegisterRoute(Route{Path: "/static", Component: func() core.Component { return routeComponent{} }})
	RegisterRoute(Route{
		Path:      "/users",
		Component: func() core.Component { return routeComponent{} },
		Children: []Route{
			{
				Path:      ":id",
				Component: func() core.Component { return routeComponent{} },
			},
			{
				Path:      ":id/profile",
				Component: func() core.Component { return routeComponent{} },
			},
		},
	})

	defs := RegisteredRoutes()
	if len(defs) != 2 {
		t.Fatalf("expected 2 top level routes, got %d", len(defs))
	}
	if defs[0].Path != "/static" || len(defs[0].Params) != 0 {
		t.Fatalf("unexpected static route: %+v", defs[0])
	}

	users := defs[1]
	if users.Path != "/users" {
		t.Fatalf("expected /users path, got %s", users.Path)
	}
	if len(users.Children) != 2 {
		t.Fatalf("expected two children, got %d", len(users.Children))
	}
	child := users.Children[0]
	if child.Path != "/users/:id" {
		t.Fatalf("expected /users/:id full path, got %s", child.Path)
	}
	if len(child.Params) != 1 || child.Params[0] != "id" {
		t.Fatalf("expected id param, got %+v", child.Params)
	}
	profile := users.Children[1]
	if profile.Path != "/users/:id/profile" {
		t.Fatalf("expected /users/:id/profile path, got %s", profile.Path)
	}
}
