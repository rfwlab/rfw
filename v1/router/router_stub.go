//go:build !js || !wasm

// Package router provides no-op stubs for non-JS/WASM builds.
package router

import "github.com/rfwlab/rfw/v1/core"

// Guard is a stubbed route guard for non-WASM builds.
type Guard func(map[string]string) bool

// Route describes a route mapping in stub builds.
type Route struct {
	Path      string
	Component func() core.Component
	Guards    []Guard
	Children  []Route
}

// RegisteredRoute describes a registered route in stub builds.
type RegisteredRoute struct {
	Template string
	Path     string
	Params   []string
	Children []RegisteredRoute
}

// NotFoundComponent is ignored in stub builds.
var NotFoundComponent func() core.Component

// NotFoundCallback is ignored in stub builds.
var NotFoundCallback func(string)

// Reset is a no-op in stub builds.
func Reset() {}

// RegisterRoute is a no-op in stub builds.
func RegisterRoute(r Route) {}

// Navigate is a no-op in stub builds.
func Navigate(fullPath string) {}

// RegisteredRoutes returns nil in stub builds.
func RegisteredRoutes() []RegisteredRoute { return nil }

// ExposeNavigate is a no-op in stub builds.
func ExposeNavigate() {}

// InitRouter is a no-op in stub builds.
func InitRouter() {}
